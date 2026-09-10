package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"campushub/backend/internal/config"
	"campushub/backend/internal/database"
	applogger "campushub/backend/internal/logger"
	"campushub/backend/internal/router"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

var (
	ginLambda *ginadapter.GinLambda
	log       *slog.Logger
)

func init() {
	cfg := config.Load()
	log = applogger.New()

	db, err := database.NewPool(database.Config{
		DSN:             cfg.DatabaseURL,
		MaxOpenConns:    cfg.DBMaxOpenConns, // kept small deliberately
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: 5 * time.Minute,
	})
	if err != nil {
		log.Error("failed to connect to database at cold start", "error", err)
		os.Exit(1)
	}

	engine := router.New(router.Dependencies{DB: db, Logger: log, JWTSecret: cfg.JWTSecret})
	ginLambda = ginadapter.New(engine)
}

func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handleRequest)
}
