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
	"medods-test/internal/mail/mockmail"
	servicev0 "medods-test/internal/service/v0"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Logger.Level(zerolog.InfoLevel)
}

func main() {
	addr := ":" + os.Getenv("PORT")
	logger := &log.Logger
	logger.Info().Str("addr", addr).Msg("starting")
	ctx := logger.WithContext(context.Background())

	ctx, stopSignal := signal.NotifyContext(ctx, os.Interrupt)
	defer stopSignal()

	db, err := pgx.Connect(ctx, os.Getenv("DB_URL")+"?sslmode=disable")
	if err != nil {
		logger.Error().Err(err).Msg("failed to connect to db")
		os.Exit(1)
	}

	secret, err := base64.StdEncoding.DecodeString(os.Getenv("SECRET"))
	if err != nil {
		panic("invalid secret")
	}

	signer := jwt.NewSigner(15*time.Minute, secret)
	api := httpapi.New(servicev0.NewService(ctx, signer, mockmail.New(), db), logger)
	server := httpapi.NewServer(api, addr)

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error().Err(err).Msg("server stopped with error")
		os.Exit(1)
	}
}
