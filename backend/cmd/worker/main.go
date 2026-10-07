package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"campushub/backend/internal/cloud"
	applogger "campushub/backend/internal/logger"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var (
	snsClient *sns.Client
	topicARN  string
	log       *slog.Logger
)

// This function deliberately has no database connection and is NOT
// VPC-attached - see the Day 14 design note. Its only job is: receive
// from SQS, forward to SNS.
func init() {
	log = applogger.New()
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Error("failed to load AWS config", "error", err)
		os.Exit(1)
	}
	snsClient = sns.NewFromConfig(cfg)
	topicARN = os.Getenv("SNS_TOPIC_ARN")
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		var event cloud.NotificationEvent
		if err := json.Unmarshal([]byte(record.Body), &event); err != nil {
			// A malformed message can't be fixed by retrying - log and
			// move on rather than failing the whole batch over one record.
			log.Error("failed to unmarshal message", "error", err, "message_id", record.MessageId)
			continue
		}

		_, err := snsClient.Publish(ctx, &sns.PublishInput{
			TopicArn: aws.String(topicARN),
			Subject:  aws.String(event.Subject),
			Message:  aws.String(event.Message),
		})
		if err != nil {
			// A real SNS failure IS worth retrying - returning the error
			// makes SQS redeliver this message later.
			log.Error("failed to publish to SNS", "error", err, "message_id", record.MessageId)
			return err
		}
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
