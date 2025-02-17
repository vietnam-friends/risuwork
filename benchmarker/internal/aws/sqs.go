package aws

import (
	"context"
	"log/slog"

	slogcontext "github.com/PumpkinSeed/slog-context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type QueueReceiver[T any] struct {
	client   *sqs.Client
	queueURL string
}

func NewQueueReceiver[T any](config aws.Config, queueURL string) *QueueReceiver[T] {
	client := sqs.NewFromConfig(config)
	return &QueueReceiver[T]{
		client:   client,
		queueURL: queueURL,
	}
}

func (q *QueueReceiver[T]) TryReceive(ctx context.Context, messageDecoder func(string) (*T, error)) (*T, bool) {
	res, err := q.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.queueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     1,
	})
	if err != nil {
		return nil, false
	}

	if len(res.Messages) == 0 {
		return nil, false
	}

	msg := res.Messages[0]
	ctx = slogcontext.WithValue(ctx, "message_id", *msg.MessageId)
	slog.DebugContext(ctx, "Received message from SQS")

	t, err := messageDecoder(*msg.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to decode message", slog.String("error", err.Error()), slog.String("body", *msg.Body))
		return nil, false
	}
	slog.DebugContext(ctx, "Successfully decoded message", slog.Any("job", t))

	// すぐデリートすることで、visibility timeoutによる重複を防ぐのと、FIFOキューから別のNodeが受け取れるようにする
	if _, err := q.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	}); err != nil {
		slog.ErrorContext(ctx, "Failed to delete message", slog.String("error", err.Error()))
		return nil, false
	}
	slog.DebugContext(ctx, "Delete message from SQS")
	return t, true
}
