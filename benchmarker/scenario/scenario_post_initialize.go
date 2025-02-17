package scenario

import (
	"context"
	"fmt"
	"log/slog"
	"risuwork-benchmarker/internal/logger"
	"risuwork-benchmarker/scenario/api"

	"github.com/isucon/isucandar"
)

type InitializeResponse struct {
	Lang string `json:"lang"`
}

func (s *Scenario) postInitialize(ctx context.Context, step *isucandar.BenchmarkStep) error {
	ag, err := NewAgent(s.Option.TargetHost, s.Option.InitializeRequestTimeout, s.Option.BenchID)
	if err != nil {
		err = fmt.Errorf("NewAgent: %w", err)
		addError(ctx, step, ErrCritical, err)
		return err
	}

	errorTag := "POST /initialize"
	resp, err := checkWithModel[InitializeResponse](api.PostInitialize(ctx, ag))
	if err != nil {
		err = fmt.Errorf("%s: %w", errorTag, err)
		addError(ctx, step, ErrCritical, err)
		return err
	}
	logger.Admin().Info(fmt.Sprintf("言語: %s", resp.Lang), slog.String("lang", resp.Lang))
	addScore(ctx, step, ScoreNormalPost)

	return nil
}
