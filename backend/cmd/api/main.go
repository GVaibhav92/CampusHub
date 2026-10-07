package main

import (
	"net/http"
	"os"
	"time"

	"campushub/backend/internal/auth"
	"campushub/backend/internal/config"
	"campushub/backend/internal/database"
	applogger "campushub/backend/internal/logger"
	"campushub/backend/internal/middleware"
	"campushub/backend/internal/repository"
	"campushub/backend/internal/router"
)

const tokenTTL = 24 * time.Hour

func main() {
	cfg := config.Load()
	log := applogger.New()

	db, err := database.NewPool(database.Config{
		DSN: cfg.DatabaseURL, MaxOpenConns: cfg.DBMaxOpenConns, MaxIdleConns: cfg.DBMaxIdleConns, ConnMaxLifetime: 5 * time.Minute,
	})
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	issuer := middleware.NewLocalTokenIssuer(cfg.JWTSecret, tokenTTL)
	validator := middleware.NewLocalTokenValidator(cfg.JWTSecret)
	provider := auth.NewLocalProvider(userRepo, issuer)

	engine := router.New(router.Dependencies{DB: db, Logger: log, AuthProvider: provider, TokenValidator: validator, Publisher: nil})

	log.Info("starting local server", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, engine); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
