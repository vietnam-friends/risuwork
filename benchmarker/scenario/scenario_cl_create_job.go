package scenario

import (
	"context"
	"fmt"
	"risuwork-benchmarker/scenario/api"
	"risuwork-benchmarker/scenario/fixture"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

func (s *Scenario) clCreateJobScenario(step *isucandar.BenchmarkStep) worker.WorkerFunc {
	return func(ctx context.Context, _ int) {
		ag := MustNewAgent(s.Option.TargetHost, s.Option.RequestTimeout, s.Option.BenchID)
		cl, ok := s.Clients.Random()
		if !ok {
			// ベンチがまだクライアントを作成していないタイミングではスキップ
			return
		}

		errorTag := "POST /api/cl/login"
		if err := check(api.PostCLLogin(ctx, ag, cl.Email, cl.Password)); err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		addScore(ctx, step, ScoreAuth)

		// 最大5個の求人を作成
		errorTag = "POST /api/cl/job"
		n := fixture.RandInt(CreateJobMinJobs, CreateJobMaxJobs)
		for i := 0; i < n; i++ {
			func() { // deferでレスポンスを閉じるために即時関数で囲む
				job := fixture.GenerateJob()
				job.Title, job.Description = fixture.GenerateJobDescription()
				job.Salary = fixture.GenerateJobSalary()
				job.Tags = fixture.GenerateRandomTags()
				resp, err := checkWithModel[CreateResponseBody](api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags))
				if err != nil {
					addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
					return
				}
				job.ID = resp.ID
				s.Jobs.Add(*job)
				addScore(ctx, step, ScoreNormalPost)
			}()
		}

		// ログアウト
		errorTag = "POST /api/cl/logout"
		if err := check(api.PostCLLogout(ctx, ag)); err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		addScore(ctx, step, ScoreAuth)
	}
}
