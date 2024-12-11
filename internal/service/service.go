package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"medods-test/internal/jwt"
	"medods-test/internal/mail"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// NewService creates a new Service.
func NewService(ctx context.Context, signer *jwt.Signer, mailer mail.Mailer, db *pgx.Conn) *Service {
	return &Service{
		ctx:    ctx,
		signer: signer,
		mailer: mailer,
		db:     db,
	}
}

// Service is a service.
type Service struct {
	ctx    context.Context
	signer *jwt.Signer
	mailer mail.Mailer
	db     *pgx.Conn
}

func (s Service) createPair(tx transaction, userID, userIP string) (*TokenPair, error) {
	pair, refresh, err := s.GenerateTokenPair(userID, userIP)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(s.ctx,
		"INSERT INTO tokens (user_id, jwt_id, refresh_token, ip) VALUES ($1, $2, $3, $4)",
		userID, refresh.ID, refresh.TokenHash, userIP)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

// Auth generates the TokenPair.
func (s *Service) Auth(userID, userIP string) (*TokenPair, error) {
	return s.createPair(s.db, userID, userIP)
}

// Refresh refreshes the Access token.
func (s *Service) Refresh(pair TokenPair) (*TokenPair, error) {
	claims, err := s.signer.Validate(pair.AccessToken)
	if err != nil {
		return nil, err
	}

	var refreshHashStr string
	var ip []byte

	txCtx, txCancel := context.WithTimeout(s.ctx, time.Minute)
	defer txCancel()
	tx, err := s.db.Begin(txCtx)
	if err != nil {
		return nil, err
	}
	err = tx.QueryRow(s.ctx,
		"SELECT refresh_token, ip FROM tokens WHERE user_id = $1 AND jwt_id = $2",
		claims.Subject, claims.ID).Scan(&refreshHashStr, &ip)
	if err != nil {
		return nil, err
	}
	refreshHash, err := base64.RawStdEncoding.DecodeString(refreshHashStr)
	if err != nil {
		return nil, err
	}
	refreshToken, err := base64.RawURLEncoding.DecodeString(pair.RefreshToken)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword(refreshHash, refreshToken); err != nil {
		return nil, err
	}

	newPair, err := s.createPair(tx, claims.Subject, claims.UserIP)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(txCtx); err != nil {
		return nil, err
	}

	if string(ip) != claims.UserIP {
		go func() {
			err = s.mailer.Send("mock@mail.com", "The IP address was changed since last login")
			if err != nil {
				// log error
			}
		}()
	}

	return newPair, nil
}

type transaction interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// TokenPair contains the Access and Refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Refresh struct {
	ID        string
	TokenHash string
}

func (s *Service) GenerateTokenPair(userID, userIP string) (*TokenPair, *Refresh, error) {
	jwtID, err := randRead(32)
	if err != nil {
		return nil, nil, err
	}
	refresh := Refresh{
		ID: base64.RawStdEncoding.EncodeToString(jwtID),
	}
	var pair TokenPair
	pair.AccessToken, err = s.signer.Create(refresh.ID, userID, userIP)
	if err != nil {
		return nil, nil, err
	}

	refreshTokenRaw, err := randRead(32)
	if err != nil {
		return nil, nil, err
	}
	pair.RefreshToken = base64.RawURLEncoding.EncodeToString(refreshTokenRaw)

	h, err := bcrypt.GenerateFromPassword(refreshTokenRaw, bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	refresh.TokenHash = base64.RawStdEncoding.EncodeToString(h)

	return &pair, &refresh, nil
}

func randRead(size uint) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
