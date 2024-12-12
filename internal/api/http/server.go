package httpapi

import (
	"net/http"
)

// NewServer creates a new http server with mux.
// Panics if the api is nil or addr is empty.
func NewServer(api *API, addr string) *http.Server {
	if api == nil || addr == "" {
		panic("NewServer: empty parameters")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /auth", api.Auth)
	mux.HandleFunc("POST /refresh", api.Refresh)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
