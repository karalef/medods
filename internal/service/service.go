package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"medods-test/internal/jwt"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// NewService creates a new Service.
func NewService(ctx context.Context, signer *jwt.Signer, db *pgx.Conn) *Service {
	return &Service{
		ctx:    ctx,
		signer: signer,
		db:     db,
	}
}

// Service is a service.
type Service struct {
	ctx    context.Context
	signer *jwt.Signer
	db     *pgx.Conn
}

// Auth generates the TokenPair.
func (s *Service) Auth(userID, userIP string) (*TokenPair, error) {
	pair, refresh, err := s.GenerateTokenPair(userID, userIP)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(s.ctx,
		"INSERT INTO tokens (user_id, jwt_id, refresh_token, ip) VALUES ($1, $2, $3, $4)",
		userID, refresh.ID, refresh.TokenHash, userIP)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Refresh refreshes the Access token.
func (s *Service) Refresh(pair TokenPair) (*TokenPair, error) {
	claims, err := s.signer.Validate(pair.AccessToken)
	if err != nil {
		return nil, err
	}

	var hash, ip []byte
	err = s.db.QueryRow(s.ctx, "SELECT refresh_token, ip FROM tokens WHERE user_id = $1", claims.UserID).Scan(&hash, &ip)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword(hash, []byte(pair.RefreshToken)); err != nil {
		return nil, err
	}

	if string(ip) != claims.UserIP {
		// send warning email
	}

	newAccessToken, err := s.signer.Create(claims.UserID, claims.UserIP)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken: newAccessToken,
	}, nil
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

func (s *Service) GenerateTokenPair(userID, userIP string) (TokenPair, Refresh, error) {
	jwtID, err := randRead(32)
	if err != nil {
		return TokenPair{}, Refresh{}, err
	}
	refresh := Refresh{
		ID: base64.RawStdEncoding.EncodeToString(jwtID),
	}
	var pair TokenPair
	pair.AccessToken, err = s.signer.Create(refresh.ID, userID, userIP)
	if err != nil {
		return TokenPair{}, Refresh{}, err
	}

	refreshTokenRaw, err := randRead(32)
	if err != nil {
		return TokenPair{}, Refresh{}, err
	}
	pair.RefreshToken = base64.RawURLEncoding.EncodeToString(refreshTokenRaw)

	h, err := bcrypt.GenerateFromPassword(refreshTokenRaw, bcrypt.DefaultCost)
	if err != nil {
		return TokenPair{}, Refresh{}, err
	}
	refresh.RefreshTokenHash = base64.RawStdEncoding.EncodeToString(h)

	return pair, refresh, nil
}

func randRead(size uint) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
