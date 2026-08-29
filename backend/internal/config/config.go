package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	DBMaxOpenConns int
	DBMaxIdleConns int
	JWTSecret      string
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://campushub:campushub@localhost:5432/campushub?sslmode=disable"),
		DBMaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 5),
		DBMaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 2),
		JWTSecret:      getEnv("JWT_SECRET", "local-dev-secret-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
