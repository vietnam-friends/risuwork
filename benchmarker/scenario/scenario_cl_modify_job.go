package scenario

import (
	"context"
	"fmt"
	"risuwork-benchmarker/scenario/api"
	"risuwork-benchmarker/scenario/fixture"
	"risuwork-benchmarker/scenario/model"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/worker"
)

func (s *Scenario) clModifyJobScenario(step *isucandar.BenchmarkStep) worker.WorkerFunc {
	return func(ctx context.Context, _ int) {
		ag := MustNewAgent(s.Option.TargetHost, s.Option.RequestTimeout, s.Option.BenchID)
		cl, ok := s.Clients.Random()
		if !ok {
			// ベンチがまだクライアントを作成していないタイミングではスキップ
			return
		}

		// ログイン
		errorTag := "POST /api/cl/login"
		if err := check(api.PostCLLogin(ctx, ag, cl.Email, cl.Password)); err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		addScore(ctx, step, ScoreAuth)

		// 求人一覧を閲覧
		errorTag = "GET /api/cl/jobs"
		jobs := make([]model.Job, 0)
		page := 0
		for {
			jobResp := getClJobs(ctx, ag, step, &page)
			if jobResp == nil || !jobResp.HasNextPage {
				break
			}
			jobs = append(jobs, jobResp.Jobs...)
			page++
		}

		// 全求人の詳細を確認
		errorTag = "GET /api/cl/job/:jobid"
		for _, job := range jobs {
			appliedCount := func() int { // deferでレスポンスを閉じるために即時関数で囲む
				resp, err := checkWithModel[model.JobWithApplication](api.GetCLJob(ctx, ag, job.ID))
				if err != nil {
					addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
					return 0
				}
				addScore(ctx, step, ScoreNormalGET)
				return len(resp.Applications)
			}()

			// 求人を編集
			errorTag = "POST /api/cl/job/:jobid/archive"
			if appliedCount > 5 {
				// アーカイブ
				func() { // deferでレスポンスを閉じるために即時関数で囲む
					if err := check(api.POSTCLJobArchive(ctx, ag, job.ID)); err != nil {
						addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
						return
					}
					addScore(ctx, step, ScoreNormalPost)
				}()
			} else {
				// 編集
				errorTag = "PATCH /api/cl/job/:jobid"
				func() { // deferでレスポンスを閉じるために即時関数で囲む
					title := job.Title
					description := job.Description
					salary := fixture.GenerateJobSalary()
					tags := fixture.GenerateRandomTags()
					isActive := true
					if err := check(api.PatchCLJob(ctx, ag, job.ID, &title, &description, &salary, &tags, &isActive)); err != nil {
						addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
						return
					}
					addScore(ctx, step, ScoreNormalPatch)
				}()
			}
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

func getClJobs(ctx context.Context, ag *agent.Agent, step *isucandar.BenchmarkStep, page *int) *model.JobsResponse {
	errorTag := "GET /api/cl/jobs"
	resp, err := checkWithModel[model.JobsResponse](api.GetCLJobs(ctx, ag, page))
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
		return nil
	}
	addScore(ctx, step, ScoreNormalGET)
	return resp
}
