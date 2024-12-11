package main

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"os/signal"
	"time"

	httpapi "medods-test/internal/api/http"
	"medods-test/internal/jwt"
	"medods-test/internal/mail"
	"medods-test/internal/service"

	"github.com/jackc/pgx/v5"
)

/*
.env

SECRET
PORT
DB_URL
*/

func main() {
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignal()

	db, err := pgx.Connect(ctx, os.Getenv("DB_URL")+"?sslmode=disable")
	if err != nil {
		panic(err)
	}

	secret, err := base64.StdEncoding.DecodeString(os.Getenv("SECRET"))
	if err != nil {
		panic(err)
	}

	signer := jwt.NewSigner(15*time.Minute, secret)
	mailer, _ := mail.NewMock()
	api := httpapi.New(service.NewService(ctx, signer, mailer, db))

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /auth", api.Auth)
	mux.HandleFunc("GET /refresh", api.Refresh)

	server := http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: mux,
	}
	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
