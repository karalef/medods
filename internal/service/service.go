package service

// Service represents the test task service.
type Service interface {
	// Auth generates the TokenPair.
	Auth(userID, userIP string) (*TokenPair, error)

	// Refresh refreshes the access token.
	Refresh(pair TokenPair) (*TokenPair, error)
}

// TokenPair contains the Access and Refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Unauthorized error type.
type Unauthorized struct{ Err error }

func (u Unauthorized) Error() string { return u.Err.Error() }
