package httpapi

import (
	"encoding/json"
	"net/http"

	"medods-test/internal/service"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// New creates new http api.
func New(srv service.Service, log *zerolog.Logger) *API {
	return &API{srv: srv, log: log}
}

type API struct {
	srv service.Service
	log *zerolog.Logger
}

// Auth returns the Access and Refresh tokens.
func (s *API) Auth(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	userID := r.URL.Query().Get("user_id")
	if _, err := uuid.Parse(userID); err != nil {
		s.log.Error().
			Str("ip", ip).
			Str("user_id", userID).
			Str("request", r.URL.Path).Msg("invalid user id")
		s.errorResponse(w, r, http.StatusBadRequest, "invalid user id")
		return
	}

	pair, err := s.srv.Auth(userID, ip)
	if err != nil {
		s.log.Error().
			Str("ip", ip).
			Str("user_id", userID).
			Str("request", r.URL.Path).Err(err).Send()
		s.errorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, r, pair)
}

// Refresh refreshes the Access token.
func (s *API) Refresh(w http.ResponseWriter, r *http.Request) {
	var pair service.TokenPair
	if err := json.NewDecoder(r.Body).Decode(&pair); err != nil {
		s.log.Error().Str("request", r.URL.Path).Err(err).Msg("failed to parse request body")
		s.errorResponse(w, r, http.StatusBadRequest, "failed to parse request body: "+err.Error())
		return
	}

	newPair, err := s.srv.Refresh(pair)
	if err != nil {
		s.log.Error().Str("request", r.URL.Path).Err(err).Msg("failed to parse request body")
		switch err.(type) {
		case service.Unauthorized:
			s.errorResponse(w, r, http.StatusUnauthorized, err.Error())
		default:
			s.errorResponse(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	s.jsonResponse(w, r, newPair)
}

func (s *API) errorResponse(w http.ResponseWriter, r *http.Request, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": err,
	}); err != nil {
		s.log.Error().Str("request", r.URL.Path).Err(err).Msg("failed to response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *API) jsonResponse(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Error().Str("request", r.URL.Path).Err(err).Msg("failed to response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getIP(r *http.Request) string {
	fwd := r.Header.Get("X-Forwarded-For")
	if fwd != "" {
		return fwd
	}
	return r.RemoteAddr
}
