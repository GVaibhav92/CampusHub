package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"campushub/backend/internal/auth"
	"campushub/backend/internal/cloud"
	"campushub/backend/internal/database"
	applogger "campushub/backend/internal/logger"
	"campushub/backend/internal/middleware"
	"campushub/backend/internal/router"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

var (
	ginLambda *ginadapter.GinLambdaV2
	log       *slog.Logger
)

type dbSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

func loadDSN(ctx context.Context, cfg aws.Config, secretARN string) (string, error) {
	client := secretsmanager.NewFromConfig(cfg)
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: aws.String(secretARN)})
	if err != nil {
		return "", fmt.Errorf("getting secret: %w", err)
	}
	var s dbSecret
	if err := json.Unmarshal([]byte(*out.SecretString), &s); err != nil {
		return "", fmt.Errorf("parsing secret: %w", err)
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require", s.Username, s.Password, s.Host, s.Port, s.DBName), nil
}

func init() {
	ctx := context.Background()
	log = applogger.New()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("failed to load AWS config", "error", err)
		os.Exit(1)
	}

	// Secrets Manager holds the real credentials - env vars here only
	// ever carry non-secret identifiers (an ARN, a pool ID).
	dsn, err := loadDSN(ctx, awsCfg, os.Getenv("DB_SECRET_ARN"))
	if err != nil {
		log.Error("failed to load DB secret", "error", err)
		os.Exit(1)
	}

	db, err := database.NewPool(database.Config{
		DSN: dsn, MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifetime: 5 * time.Minute, // kept small deliberately
	})
	if err != nil {
		log.Error("failed to connect to database at cold start", "error", err)
		os.Exit(1)
	}

	poolID := os.Getenv("COGNITO_USER_POOL_ID")
	clientID := os.Getenv("COGNITO_APP_CLIENT_ID")
	region := os.Getenv("AWS_REGION") // Lambda sets this automatically, never override it

	provider := auth.NewCognitoProvider(cognitoidentityprovider.NewFromConfig(awsCfg), poolID, clientID)
	validator, err := middleware.NewCognitoTokenValidator(region, poolID, clientID)
	if err != nil {
		log.Error("failed to initialize Cognito token validator", "error", err)
		os.Exit(1)
	}

	queueURL := os.Getenv("SQS_QUEUE_URL")
	publisher := cloud.NewEventPublisher(sqs.NewFromConfig(awsCfg), queueURL)

	engine := router.New(router.Dependencies{
		DB: db, Logger: log, AuthProvider: provider, TokenValidator: validator, Publisher: publisher,
	})
	ginLambda = ginadapter.NewV2(engine)
}

func handleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handleRequest)
}
