package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"risuwork-benchmarker/internal"
	internalAws "risuwork-benchmarker/internal/aws"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/PumpkinSeed/slog-context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-errors/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var execPath string

func init() {
	ep, err := os.Executable()
	if err != nil {
		panic(err)
	}
	execPath = ep
}

func SuperviseCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supervise",
		Short: "Supervise",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			recvQueueURL := v.GetString("queue-url")
			topicArn := v.GetString("topic-arn")
			tableName := v.GetString("table-name")
			slog.Debug("options", slog.String("queue-url", recvQueueURL), slog.String("topic-arn", topicArn), slog.String("table-name", tableName))
			opt := &supervisorOption{QueueURL: recvQueueURL, TopicARN: topicArn, TableName: tableName}
			return RunSupervise(cmd.Context(), opt)
		},
	}

	cmd.Flags().String("queue-url", "", "SQS queue URL")
	cmd.Flags().String("topic-arn", "", "SNS topic ARN")
	cmd.Flags().String("table-name", "", "DynamoDB table name")
	_ = v.BindPFlags(cmd.Flags())

	return cmd
}

type supervisorOption struct {
	QueueURL  string
	TopicARN  string
	TableName string
}

func RunSupervise(ctx context.Context, opt *supervisorOption) error {
	slog.InfoContext(ctx, "Start Supervisor...")

	ctx, cancel := internal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// cfg, err := config.LoadDefaultConfig(ctx, config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody))
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	sqsClient := internalAws.NewQueueReceiver[Job](cfg, opt.QueueURL)
	snsClient := internalAws.NewNotifier(cfg, opt.TopicARN)
	dynamoClient := internalAws.NewBenchResultClient(cfg, opt.TableName)

	slog.DebugContext(ctx, "Start receiving jobs from SQS")
loop:
	for {
		select {
		case <-ctx.Done():
			return extractContextError(ctx)
		default:
			// SQSからジョブを取得しようとする
			job, ok := sqsClient.TryReceive(ctx, parseJob)
			if !ok {
				time.Sleep(1 * time.Second) // sqsClient.TryReceiveのWaitTimeSecondsもあるが、呼び出し元で明示的にsleepすることでSQSに負荷をかけない
				continue loop
			}

			{ // ベンチ実行の文脈でctxのスコープを絞るために、明示的にブロックを作る
				ctx := slogcontext.WithValue(ctx, "job_id", job.ID)
				ctx = slogcontext.WithValue(ctx, "team", job.Team)
				ctx = slogcontext.WithValue(ctx, "endpoint", job.Endpoint)

				// ベンチ実行前にDynamoDBに初期値のレコードを書き込む
				startedAt := time.Now().Format(time.RFC3339Nano)
				if err := dynamoClient.Start(ctx, job.ID, startedAt); err != nil {
					return err
				}

				// ベンチ実行
				res, err := ExecBench(ctx, job)
				if err != nil {
					continue loop // supervisorがKillされたりベンチ全体がタイムアウトした場合、意図しない結果で上書きしてしまう可能性があるためDynamoDBに結果は保存しない
				}
				// ベンチが正常に終了した場合もスコアが0点で終了した場合も同様に扱い結果を更新する
				slog.InfoContext(ctx, "Finish benchmark job", slog.Any("result", res))
				endedAt := time.Now().Format(time.RFC3339Nano)
				msgs := make([]string, 0, len(res.Messages))
				for _, msg := range res.Messages {
					msgs = append(msgs, msg.Msg)
				}
				if err := dynamoClient.End(ctx, job.ID, res.Score, res.Lang, msgs, endedAt); err != nil {
					return err
				}
				if err := snsClient.NotifyResult(ctx, job.Team, job.ID, job.Endpoint, res.Score, res.Lang, msgs, job.Commit, job.Actor, job.QueuedAt, startedAt, endedAt); err != nil {
					return err
				}
			}
		}
	}
}

type Job struct {
	ID       string `json:"id"`
	Team     string `json:"team"`
	Endpoint string `json:"endpoint"`
	Commit   string `json:"commit"`
	Actor    string `json:"actor"`
	QueuedAt string `json:"queued_at"`
}

func parseJob(body string) (*Job, error) {
	var job *Job
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&job); err != nil {
		return nil, err
	}
	return job, nil
}

type BenchResult struct {
	Score    int64
	Lang     string
	Messages []LogMsg
}

type LogMsg struct {
	Timestamp string
	Msg       string
}

func ExecBench(ctx context.Context, job *Job) (*BenchResult, error) {
	// 余裕を見て3分
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	slog.InfoContext(ctx, "Start benchmark job", slog.Any("job", job))

	benchOptions := []string{"run", job.Endpoint, "--prepare-only=false", "--bench-id", job.ID}
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, execPath, benchOptions...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	cmd.WaitDelay = 13 * time.Second

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if err := cmd.Run(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil, extractContextError(ctx)
	case err, ok := <-errCh: // errCh <- err か close(errCh) までブロック
		if ok && err != nil {
			// Benchmark failed
			slog.ErrorContext(ctx, "Failed to run benchmark", slog.String("error", err.Error()))
		}
	}

	result := &BenchResult{Messages: make([]LogMsg, 0)}
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		var logJson map[string]interface{}
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		decoder.UseNumber()
		if err := decoder.Decode(&logJson); err != nil {
			slog.WarnContext(ctx, "Failed to decode bench log output", slog.String("error", err.Error()))
			continue
		}

		if score, ok := logJson["score"]; ok {
			slog.InfoContext(ctx, "Successfully get score", slog.Any("score", score))
			score, err := score.(json.Number).Int64()
			if err != nil {
				slog.WarnContext(ctx, "Failed to parse score", slog.String("error", err.Error()))
				continue
			}
			result.Score = score
		}

		if lang, ok := logJson["lang"]; ok {
			slog.InfoContext(ctx, "Successfully get lang", slog.Any("lang", lang))
			result.Lang = lang.(string)
		}

		timestamp, okTimestamp := logJson["time"]
		msg, okMsg := logJson["msg"]
		logFor, okFor := logJson["for"]
		if okTimestamp && okMsg && okFor {
			slog.InfoContext(ctx, msg.(string), slog.Any("raw", logJson))
			if logFor.(string) == "player" {
				// Player向けのログはPlayerが閲覧できるように結果としてDynamoDBに書き込む
				result.Messages = append(result.Messages, LogMsg{Timestamp: timestamp.(string), Msg: msg.(string)})
			}
		}
	}

	slices.SortStableFunc(result.Messages, func(a, b LogMsg) int {
		return strings.Compare(a.Msg, b.Msg)
	})
	result.Messages = slices.CompactFunc(result.Messages, func(a, b LogMsg) bool {
		return a.Msg == b.Msg
	})
	slices.SortStableFunc(result.Messages, func(a, b LogMsg) int {
		return strings.Compare(a.Timestamp, b.Timestamp)
	})
	return result, nil
}

func extractContextError(ctx context.Context) error {
	if cause := context.Cause(ctx); cause != nil {
		return errors.Wrap(fmt.Errorf("%s: %w", ctx.Err(), cause), 1)
	}
	return errors.Wrap(ctx.Err(), 1)
}
