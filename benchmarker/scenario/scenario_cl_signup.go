package scenario

import (
	"context"
	"fmt"
	"risuwork-benchmarker/scenario/api"
	"risuwork-benchmarker/scenario/fixture"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

func (s *Scenario) clSignupScenario(step *isucandar.BenchmarkStep) worker.WorkerFunc {
	return func(ctx context.Context, _ int) {
		ag := MustNewAgent(s.Option.TargetHost, s.Option.RequestTimeout, s.Option.BenchID)

		errorTag := "POST /api/cl/signup"
		cl := fixture.GenerateCLUser()
		company, ok := s.Companies.Random()
		if !ok {
			// ベンチがまだ会社を作成していないタイミングではスキップ
			return
		}
		resp, err := checkWithModel[CreateResponseBody](api.PostCLSignup(ctx, ag, cl.Email, cl.Password, cl.Name, company.ID))
		if err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		cl.ID = resp.ID
		s.Clients.Add(*cl)
		addScore(ctx, step, ScoreAuth)

		// ログイン
		errorTag = "POST /api/cl/login"
		if err := check(api.PostCLLogin(ctx, ag, cl.Email, cl.Password)); err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		addScore(ctx, step, ScoreAuth)

		// 求人一覧を閲覧
		errorTag = "GET /api/cl/jobs"
		if err := check(api.GetCLJobs(ctx, ag, nil)); err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		addScore(ctx, step, ScoreNormalGET)

		// ログアウト
		errorTag = "POST /api/cl/logout"
		if err := check(api.PostCLLogout(ctx, ag)); err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		addScore(ctx, step, ScoreAuth)
	}
}
