package aws

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type BenchResultClient struct {
	client    *dynamodb.Client
	tableName string
}

func NewBenchResultClient(config aws.Config, tableName string) *BenchResultClient {
	client := dynamodb.NewFromConfig(config)
	return &BenchResultClient{
		client:    client,
		tableName: tableName,
	}
}

func (b *BenchResultClient) Start(ctx context.Context, id, startedAt string) error {
	res, err := b.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(b.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: id,
			},
		},
		UpdateExpression: aws.String("SET bench_status = :bench_status, started_at = :started_at"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":bench_status": &types.AttributeValueMemberS{
				Value: "running",
			},
			":started_at": &types.AttributeValueMemberS{
				Value: startedAt,
			},
		},
	})
	slog.DebugContext(ctx, "DynamoDB create result", "response", res, "error", err)
	return err
}

func (b *BenchResultClient) End(ctx context.Context, id string, score int64, lang string, messages []string, endedAt string) error {
	status := "done"
	if score == 0 {
		status = "failed"
	}
	res, err := b.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(b.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{
				Value: id,
			},
		},
		UpdateExpression: aws.String("SET bench_status = :bench_status, score = :score, lang = :lang, messages = :messages, ended_at = :ended_at"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":bench_status": &types.AttributeValueMemberS{
				Value: status,
			},
			":score": &types.AttributeValueMemberN{
				Value: fmt.Sprintf("%d", score),
			},
			":lang": &types.AttributeValueMemberS{
				Value: lang,
			},
			":messages": &types.AttributeValueMemberS{
				Value: strings.Join(messages, "\n"),
			},
			":ended_at": &types.AttributeValueMemberS{
				Value: endedAt,
			},
		},
	})
	slog.DebugContext(ctx, "DynamoDB update result", "response", res, "error", err)
	return err
}
