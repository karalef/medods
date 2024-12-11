package httpapi

import (
	"encoding/json"
	"net/http"

	"medods-test/internal/service"
)

func New(srv *service.Service) *API {
	return &API{srv: srv}
}

type API struct {
	srv *service.Service
}

// Auth returns the Access and Refresh tokens.
func (s *API) Auth(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := q.Get("user_id")
	ip := getIP(r)

	pair, err := s.srv.Auth(userID, ip)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, pair)
}

// Refresh refreshes the Access token.
func (s *API) Refresh(w http.ResponseWriter, r *http.Request) {
	var pair service.TokenPair
	if err := json.NewDecoder(r.Body).Decode(&pair); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newPair, err := s.srv.Refresh(pair)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsonResponse(w, newPair)
}

func errorResponse(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": err,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
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
