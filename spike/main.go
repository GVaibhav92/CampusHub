package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Response struct {
	Message string `json:"message"`
	Version string `json:"postgres_version,omitempty"`
	Error   string `json:"error,omitempty"`
}

func handler(ctx context.Context) (Response, error) {
	dsn := os.Getenv("DATABASE_URL")
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return Response{Error: err.Error()}, nil
	}
	defer db.Close()

	var version string
	if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		return Response{Error: err.Error()}, nil
	}
	return Response{Message: "connected successfully", Version: version}, nil
}

func main() {
	lambda.Start(handler)
}
