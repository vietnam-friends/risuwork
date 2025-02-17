package scenario

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"risuwork-benchmarker/scenario/api"
	"risuwork-benchmarker/scenario/fixture"
	"risuwork-benchmarker/scenario/model"
	"risuwork-benchmarker/scenario/validate"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/worker"
)

func (s *Scenario) csApplyJobsScenario(step *isucandar.BenchmarkStep) worker.WorkerFunc {
	return func(ctx context.Context, _ int) {
		ag := MustNewAgent(s.Option.TargetHost, s.Option.RequestTimeout, s.Option.BenchID)

		// 25%で新規ユーザー、75%で既存ユーザーと使い分ける
		cs, ok := s.Customers.Random()
		if !ok || ApplyJobsNewCustomerRate <= rand.Float32() {
			// 新規ユーザー
			errorTag := "POST /api/cs/signup"
			cs = fixture.GenerateCSUser()
			resp, err := checkWithModel[CreateResponseBody](api.PostCSSignup(ctx, ag, cs.Email, cs.Password, cs.Name))
			if err != nil {
				addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
				return
			}
			cs.ID = resp.ID
			s.Customers.Add(*cs)
			addScore(ctx, step, ScoreAuth)
		}

		// ログイン
		errorTag := "POST /api/cs/login"
		if err := check(api.PostCSLogin(ctx, ag, cs.Email, cs.Password)); err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		addScore(ctx, step, ScoreAuth)

		// 求人検索〜応募を一定確率で5〜10回繰り返す
		n := fixture.RandInt(ApplyJobsMinSearchApplyLoop, ApplyJobsMaxSearchApplyLoop)
		for i := 0; i < n; i++ {
			func() {
				// 求人検索
				errorTag = "GET /api/cs/job/search"
				minSalaries := []int{5_000_000, 7_500_000, 10_000_000, 12_000_0000}
				q := api.JobSearchQuery{
					// 検索は単一タグのみ
					Tag:       url.QueryEscape(fixture.GenerateRandomTagsN(1)),
					MinSalary: minSalaries[rand.Intn(len(minSalaries))],
				}
				searchResp, err := checkWithModel[model.JobSearchResponse](api.GetCSJobSearch(ctx, ag, q))
				if err != nil {
					addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
					return
				}
				addScore(ctx, step, ScoreNormalGET)
				jobs := searchResp.Jobs
				if len(jobs) == 0 {
					return
				}
				var jobID int
				rand.Shuffle(len(jobs), func(i, j int) { jobs[i], jobs[j] = jobs[j], jobs[i] })
				for _, job := range jobs {
					if !s.Applications.Exists(model.Application{UserID: cs.ID, JobID: job.ID}) {
						jobID = job.ID
					}
				}

				// 応募
				errorTag = "POST /api/cs/application"
				if err := check(api.PostCSApplication(ctx, ag, jobID)); err != nil {
					var sErr validate.StatusCodeUnMatchError
					if errors.As(err, &sErr) && sErr.Actual == http.StatusConflict {
						return // NOTE: 並列で同時応募の場合にエラーが発生してしまうが、エラーとして扱う必要はない
					}
					addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
					return
				}
				s.Applications.Add(model.Application{UserID: cs.ID, JobID: jobID})
				addScore(ctx, step, ScoreJobApplication)
			}()
		}

		// 応募一覧を閲覧
		errorTag = "GET /api/cs/applications"
		page := 0
		for {
			appResp := getCSApplications(ctx, ag, step, &page)
			if appResp != nil && !appResp.HasNextPage {
				break
			}
			page++
		}

		// ログアウト
		errorTag = "POST /api/cs/logout"
		if err := check(api.PostCSLogout(ctx, ag)); err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		addScore(ctx, step, ScoreAuth)
	}
}

func getCSApplications(ctx context.Context, ag *agent.Agent, step *isucandar.BenchmarkStep, page *int) *model.ApplicationsResponse {
	errorTag := "GET /api/cs/applications"

	resp, err := checkWithModel[model.ApplicationsResponse](api.GetCSApplications(ctx, ag, page))
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
		return nil
	}
	addScore(ctx, step, ScoreNormalGET)
	return resp
}
