package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Event struct {
	Action string `json:"action"`
}

type Response struct {
	Action         string `json:"action"`
	Message        string `json:"message"`
	MigrationTable bool   `json:"migration_table_exists,omitempty"`
	Version        uint   `json:"version"`
	Dirty          bool   `json:"dirty"`
}

type DBSecret struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func loadDatabaseURL(ctx context.Context) (string, error) {
	secretARN := os.Getenv("DB_SECRET_ARN")
	if secretARN == "" {
		return "", errors.New("DB_SECRET_ARN is not configured")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("load AWS configuration: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretARN,
	})
	if err != nil {
		return "", fmt.Errorf("retrieve database secret: %w", err)
	}
	if result.SecretString == nil {
		return "", errors.New("database secret has no SecretString")
	}

	var secret DBSecret
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		return "", fmt.Errorf("parse database secret: %w", err)
	}
	if secret.Host == "" || secret.DBName == "" ||
		secret.Username == "" || secret.Password == "" {
		return "", errors.New("database secret is missing required fields")
	}
	if secret.Port == 0 {
		secret.Port = 5432
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(secret.Username, secret.Password),
		Host:   net.JoinHostPort(secret.Host, strconv.Itoa(secret.Port)),
		Path:   "/" + secret.DBName,
	}
	q := url.Values{}
	q.Set("sslmode", "require")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func handler(ctx context.Context, event Event) (Response, error) {
	dsn, err := loadDatabaseURL(ctx)
	if err != nil {
		return Response{}, err
	}

	switch strings.ToLower(strings.TrimSpace(event.Action)) {
	case "inspect":
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return Response{}, fmt.Errorf("open database: %w", err)
		}
		defer db.Close()

		if err := db.PingContext(ctx); err != nil {
			return Response{}, fmt.Errorf("connect to database: %w", err)
		}

		var tableName sql.NullString
		err = db.QueryRowContext(
			ctx,
			"SELECT to_regclass('public.schema_migrations')::text",
		).Scan(&tableName)
		if err != nil {
			return Response{}, fmt.Errorf("inspect migration table: %w", err)
		}

		if !tableName.Valid {
			return Response{
				Action:         "inspect",
				Message:        "Connected successfully; migration table does not exist.",
				MigrationTable: false,
			}, nil
		}

		var version int64
		var dirty bool
		err = db.QueryRowContext(
			ctx,
			"SELECT version, dirty FROM public.schema_migrations LIMIT 1",
		).Scan(&version, &dirty)
		if err != nil {
			return Response{}, fmt.Errorf("read migration version: %w", err)
		}

		return Response{
			Action:         "inspect",
			Message:        "Connected successfully; migration state read.",
			MigrationTable: true,
			Version:        uint(version),
			Dirty:          dirty,
		}, nil

	case "up":
		if strings.ToLower(os.Getenv("ALLOW_MIGRATIONS")) != "true" {
			return Response{}, errors.New(
				"migration execution is disabled; ALLOW_MIGRATIONS must be true",
			)
		}

		m, err := migrate.New(
			"file:///var/task/migrations",
			dsn,
		)
		if err != nil {
			return Response{}, fmt.Errorf("initialize migrations: %w", err)
		}
		defer func() {
			_, _ = m.Close()
		}()

		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return Response{}, fmt.Errorf("apply migrations: %w", err)
		}

		version, dirty, err := m.Version()
		if err != nil {
			return Response{}, fmt.Errorf("read final migration state: %w", err)
		}

		return Response{
			Action:         "up",
			Message:        "Migration run completed.",
			MigrationTable: true,
			Version:        version,
			Dirty:          dirty,
		}, nil

	default:
		return Response{}, errors.New("unsupported action; use inspect or up")
	}
}

func main() {
	lambda.Start(handler)
}
