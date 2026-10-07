package cloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type EventPublisher struct {
	client   *sqs.Client
	queueURL string
}

func NewEventPublisher(client *sqs.Client, queueURL string) *EventPublisher {
	return &EventPublisher{client: client, queueURL: queueURL}
}

// Publish is intentionally the only method here - one job, done once.
// Callers treat a returned error as log-and-continue, never as a reason
// to fail the request that triggered it (see booking_service.go).
func (p *EventPublisher) Publish(ctx context.Context, event NotificationEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}
	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("sending SQS message: %w", err)
	}
	return nil
}
