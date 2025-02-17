package scenario

import (
	"context"
	"fmt"
	"risuwork-benchmarker/scenario/api"
	"risuwork-benchmarker/scenario/fixture"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

func (s *Scenario) clCreateCompanyScenario(step *isucandar.BenchmarkStep) worker.WorkerFunc {
	return func(ctx context.Context, _ int) {
		ag := MustNewAgent(s.Option.TargetHost, s.Option.RequestTimeout, s.Option.BenchID)

		errorTag := "POST /api/cl/company"
		company := fixture.GenerateCompany()
		resp, err := checkWithModel[CreateResponseBody](api.PostCLCompany(ctx, ag, company.Name, company.IndustryID))
		if err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
			return
		}
		company.ID = resp.ID
		s.Companies.Add(*company)
		addScore(ctx, step, ScoreNormalPost)

		// 作った会社の担当者を最大3人追加（途中でエラーが起きた場合そこまで）
		errorTag = "POST /api/cl/signup"
		n := fixture.RandInt(CreateCompanyMinClients, CreateCompanyMaxClients)
		for i := 0; i < n; i++ {
			func() { // deferでレスポンスを閉じるために即時関数で囲む
				cl := fixture.GenerateCLUser()
				resp, err := checkWithModel[CreateResponseBody](api.PostCLSignup(ctx, ag, cl.Email, cl.Password, cl.Name, company.ID))
				if err != nil {
					addError(ctx, step, ErrCritical, fmt.Errorf("%s: %w", errorTag, err))
					return
				}
				cl.ID = resp.ID
				s.Clients.Add(*cl)
				addScore(ctx, step, ScoreAuth)
			}()
		}
	}
}
