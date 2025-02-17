package aws

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type Notifier struct {
	client   *sns.Client
	topicArn string
}

func NewNotifier(config aws.Config, topicArn string) *Notifier {
	client := sns.NewFromConfig(config)
	return &Notifier{
		client:   client,
		topicArn: topicArn,
	}
}

type benchSnsMessage struct {
	Team      string   `json:"team"`
	ID        string   `json:"id"`
	Endpoint  string   `json:"endpoint"`
	Score     int64    `json:"score"`
	Lang      string   `json:"lang"`
	Messages  []string `json:"messages"`
	Commit    string   `json:"commit"`
	Actor     string   `json:"actor"`
	QueuedAt  string   `json:"queued_at"`
	StartedAt string   `json:"started_at"`
	EndedAt   string   `json:"ended_at"`
}

func (n *Notifier) NotifyResult(ctx context.Context, team, id, endpoint string, score int64, lang string, messages []string, commit, actor, queuedAt, startedAt, endedAt string) error {
	message := benchSnsMessage{
		Team:      team,
		ID:        id,
		Endpoint:  endpoint,
		Score:     score,
		Lang:      lang,
		Messages:  messages,
		Commit:    commit,
		Actor:     actor,
		QueuedAt:  queuedAt,
		StartedAt: startedAt,
		EndedAt:   endedAt,
	}
	messageJson, err := json.Marshal(&message)
	if err != nil {
		return err
	}
	res, err := n.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(n.topicArn),
		Message:  aws.String(string(messageJson)),
	})
	slog.DebugContext(ctx, "SNS publish result", "response", res, "error", err)
	return err
}

func (n *Notifier) NotifyError(ctx context.Context, team string, reason error) error {
	_, err := n.client.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(n.topicArn),
		Message:  aws.String(reason.Error()),
	})
	return err
}
