package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// NewSigner creates a new JWT signer with HS512 algorithm.
// Panics if secret is empty.
// If expiry is less than 1 second, it is set to 1 minute.
func NewSigner(expiry time.Duration, secret []byte) *Signer {
	if len(secret) == 0 {
		panic("empty secret")
	}
	if expiry < time.Second {
		expiry = time.Minute
	}
	return &Signer{
		method: jwt.SigningMethodHS512,
		secret: secret,
		expiry: expiry,
	}
}

// Signer is a JWT signer.
type Signer struct {
	method jwt.SigningMethod
	secret []byte
	expiry time.Duration
}

// Claims is a custom JWT claims that contains user ID and IP.
type Claims struct {
	jwt.RegisteredClaims

	UserIP string `json:"uip"`
}

func (s Signer) key(_ *jwt.Token) (any, error) { return s.secret, nil }

// Create creates a new JWT token.
func (s Signer) Create(id, userID, userIP string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserIP: userIP,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
		},
	}
	return jwt.NewWithClaims(s.method, claims).SignedString(s.secret)
}

// Validate validates a JWT token.
func (s Signer) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, s.key)
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
