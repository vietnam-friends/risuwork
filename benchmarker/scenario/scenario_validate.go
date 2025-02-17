package scenario

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"risuwork-benchmarker/scenario/action"
	"risuwork-benchmarker/scenario/api"
	"risuwork-benchmarker/scenario/fixture"
	"risuwork-benchmarker/scenario/model"
	"risuwork-benchmarker/scenario/validate"
	"strings"
	"sync"

	"github.com/isucon/isucandar"
)

// すべてのエンドポイントをレスポンスを含めて詳細に検証する
func verify(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	type verifyFunc func(context.Context, *isucandar.BenchmarkStep, Option)

	verifyFuncs := []verifyFunc{
		// POST /cs/signup
		verifyCSSignupOK,
		verifyCSSignupDuplicateUser,
		// POST /cs/login
		verifyCSLoginOK,
		verifyCSLoginNotExist,
		verifyCSLoginWrongPassword,
		// POST /cs/logout
		verifyCSLogoutOK,
		verifyCSLogoutNotLogin,
		// GET /cs/job_search
		verifyCSJobSearchKeywordOK,
		verifyCSJobSearchSalaryOK,
		verifyCSJobSearchTagsOK,
		verifyCSJobSearchIndustryOK,
		// POST /cs/application
		verifyCSApplicationOK,
		verifyCSApplicationNotLogin,
		verifyCSApplicationNotExistJobID,
		verifyCSApplicationNotActiveJob,
		// GET /cs/applications
		verifyCSApplicationsOK,
		verifyCSApplicationListPagingOK,
		verifyCSApplicationsNotLogin,
		// POST /cl/company
		verifyCLCompanyOK,
		// POST /cl/signup
		verifyCLSignupOK,
		verifyCLSignupDuplicateUser,
		verifyCLSignupNotExistComapnyID,
		// POST /cl/login
		verifyCLLoginOK,
		verifyCLLoginNotExist,
		verifyCLLoginWrongPassword,
		// POST /cl/logout
		verifyCLLogoutOK,
		verifyCLLogoutNotLogin,
		// POST /cl/job
		verifyCreateJobOK,
		verifyCreateJobNotLogin,
		// PATCH /cl/job/:jobid
		verifyUpdateJobOK,
		verifyUpdateJobNotLogin,
		verifyUpdateJobNotOwnerUser,
		verifyUpdateJobNotExistJobID,
		verifyUpdateJobArchivedJob,
		// POST /cl/job/:jobid/archive
		verifyArchiveJobOK,
		verifyArchiveJobNotLogin,
		verifyArchiveJobNotOwnerUser,
		verifyArchiveJobNotExistJobID,
		// GET /cl/job/:jobid
		verifyGetJobOK,
		verifyGetJobArchivedJob,
		verifyGetJobNotLogin,
		verifyGetJobNotOwnerUser,
		verifyGetJobNotExistJobID,
		// GET /cl/jobs
		verifyGetJobsOK,
		verifyGetJobsPagingOK,
		verifyGetJobsNotLogin,
	}

	// 全てのverifyFuncを並列実行し、waitGroupで待機
	wg := &sync.WaitGroup{}
	for _, f := range verifyFuncs {
		wg.Add(1)
		f := f
		go func() {
			defer wg.Done()
			f(ctx, step, opt)
		}()
	}
	wg.Wait()
}

// POST /cs/signup 正常系
func verifyCSSignupOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cs := fixture.GenerateCSUser()

	/* Act */
	resp, err := api.PostCSSignup(ctx, ag, cs.Email, cs.Password, cs.Name)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
	type CSUser struct {
		ID int `json:"id"`
	}
	var csUser CSUser
	if err := json.NewDecoder(resp.Body).Decode(&csUser); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録APIのレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}
	if csUser.ID <= 0 {
		addError(ctx, step, ErrCritical, fmt.Errorf("登録したCSアカウントのIDが不正です: %d", csUser.ID))
		return
	}
}

// POST /cs/signup 異常系 同一ユーザーでの登録エラー
func verifyCSSignupDuplicateUser(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cs, err := action.CreateCSUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	// 同一ユーザーで登録
	resp, err := api.PostCSSignup(ctx, ag, cs.Email, cs.Password, cs.Name)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	// 同一ユーザーでの登録なので409エラーが期待値
	if err := validate.StatusCode(resp, http.StatusConflict); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

}

// POST /cs/login 正常系
func verifyCSLoginOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cs, err := action.CreateCSUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.PostCSLogin(ctx, ag, cs.Email, cs.Password)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントのログインに失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

}

// POST /cs/login 異常系 存在しないアカウント
func verifyCSLoginNotExist(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cs := fixture.GenerateCSUser()

	/* Act */
	resp, err := api.PostCSLogin(ctx, ag, cs.Email, cs.Password)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントのログインに失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

}

// POST /cs/login 異常系 誤ったパスワード
func verifyCSLoginWrongPassword(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cs, err := action.CreateCSUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp2, err := api.PostCSLogin(ctx, ag, cs.Email, "wrong"+cs.Password)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントのログインに失敗しました: %w", err))
		return
	}
	defer resp2.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp2, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

}

// POST /cs/logout 正常系
func verifyCSLogoutOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, err = action.CreateCSUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.PostCSLogout(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントのログアウトに失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

}

// POST /cs/logout 異常系 未ログインでのログアウト
func verifyCSLogoutNotLogin(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	/* Act */
	resp, err := api.PostCSLogout(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントのログアウトに失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

}

// GET /cs/job_search 正常系 keyword 検索
func verifyCSJobSearchKeywordOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// CLユーザーで求人を作成
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
	}

	cl, err := action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	keyword := fixture.GenerateRandomString()

	// titleにkeywordを含む求人を作成
	jobByTitle, err := action.CreateJobByTitle(ctx, ag, keyword+" title")
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}
	jobByTitle.Company = cl.Company

	// descriptionにkeywordを含む求人を作成
	jobByDescription, err := action.CreateJobByDescription(ctx, ag, "description "+keyword)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}
	jobByDescription.Company = cl.Company

	// keywordを含まない求人を作成
	_, err = action.CreateJob(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}

	// CSユーザー
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	/* Act */
	query := api.JobSearchQuery{
		Keyword: keyword,
	}
	resp, err := api.GetCSJobSearch(ctx, ag2, query)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	result := model.JobSearchResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索結果jsonの読み取りに失敗しました: %w", err))
		return
	}

	if len(result.Jobs) != 2 {
		addError(ctx, step, ErrCritical, fmt.Errorf("keyword検索で取得できる求人の数が期待値(2)と異なります: %d", len(result.Jobs)))
		return
	}

	err = validate.JobResultOrder(result.Jobs)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索で取得できる求人の順序が正しくありません: %w", err))
		return
	}

	// 求人の内容を確認
	// 並び順がupdated_atの降順のため、作成した逆順に確認
	err = validate.JobForCS(result.Jobs[0], *jobByDescription)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索で取得できる求人の内容が期待値と異なります: %w", err))
		return
	}
	err = validate.JobForCS(result.Jobs[1], *jobByTitle)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索で取得できる求人の内容が期待値と異なります: %w", err))
		return
	}
}

// GET /cs/job_search 正常系 salary 検索
func verifyCSJobSearchSalaryOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// CLユーザーで求人を作成
	ag1, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag1)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	// 9999円の求人を作成
	_, err = action.CreateJobBySalary(ctx, ag1, 9999)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}

	// 10000円の求人を作成
	_, err = action.CreateJobBySalary(ctx, ag1, 10000)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}

	// 19999円の求人を作成
	_, err = action.CreateJobBySalary(ctx, ag1, 19999)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}

	// 20000円の求人を作成
	_, err = action.CreateJobBySalary(ctx, ag1, 20000)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}

	// CSユーザー
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
	}

	/* Act */
	query := api.JobSearchQuery{
		MinSalary: 10000,
		MaxSalary: 20000,
	}
	resp, err := api.GetCSJobSearch(ctx, ag2, query)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	result := model.JobSearchResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索結果jsonの読み取りに失敗しました: %w", err))
		return
	}

	err = validate.JobResultOrder(result.Jobs)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索で取得できる求人の順序が正しくありません: %w", err))
		return
	}

	for _, job := range result.Jobs {
		if job.Salary < query.MinSalary {
			addError(ctx, step, ErrCritical, fmt.Errorf("min_salaryより給与の低い求人が検索結果に含まれています (job.Salary: %d, query.MinSalary: %d)", job.Salary, query.MinSalary))
			return
		}
		if job.Salary > query.MaxSalary {
			addError(ctx, step, ErrCritical, fmt.Errorf("給与がmax_salary以上の求人が検索結果に含まれています (job.Salary: %d, query.MaxSalary: %d)", job.Salary, query.MaxSalary))
			return
		}
	}
}

// GET /cs/job_search 正常系 tag 検索
func verifyCSJobSearchTagsOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// CLユーザーで求人を作成
	ag1, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag1)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	tag := fixture.GenerateRandomString()

	// 15 x 4 = 60 の求人を作成
	for i := 0; i < 15; i++ {
		_, err = action.Create4JobsByTag(ctx, ag1, tag)
		if err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
			return
		}
	}

	// CSユーザー
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	/* Act */
	// 1ページ目(50件)
	query1 := api.JobSearchQuery{
		Tag: tag,
	}
	res1, err := api.GetCSJobSearch(ctx, ag2, query1)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索に失敗しました: %w", err))
		return
	}
	defer res1.Body.Close()

	result1 := model.JobSearchResponse{}
	if err := json.NewDecoder(res1.Body).Decode(&result1); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索結果jsonの読み取りに失敗しました: %w", err))
		return
	}

	// 2ページ目(10件)
	page := 1
	query2 := api.JobSearchQuery{
		Tag:  tag,
		Page: &page,
	}
	resp2, err := api.GetCSJobSearch(ctx, ag2, query2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索に失敗しました: %w", err))
		return
	}
	defer resp2.Body.Close()

	result2 := model.JobSearchResponse{}
	if err := json.NewDecoder(resp2.Body).Decode(&result2); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索結果jsonの読み取りに失敗しました: %w", err))
		return
	}

	/* Assert */
	if err := validate.StatusCode(res1, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	// 1ページ目の件数が50件となっていること
	if len(result1.Jobs) != 50 {
		addError(ctx, step, ErrCritical, fmt.Errorf("tag検索で1ページ目に取得できる求人の数が期待値(50)と異なります: %d", len(result1.Jobs)))
		return
	}

	// 2ページ目が存在することが示されていること
	if !result1.HasNextPage {
		err := errors.New("求人検索の検索結果が51件以上にも関わらずhas_next_page が false になっています")
		addError(ctx, step, ErrCritical, err)
		return
	}

	// ソート順が正しいこと
	err = validate.JobResultOrder(result1.Jobs)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索で取得できる求人の順序が正しくありません: %w", err))
		return
	}

	// 検索時に指定したタグを含んでいること
	for _, job := range result1.Jobs {
		if !strings.Contains(job.Tags, tag) {
			addError(ctx, step, ErrCritical, fmt.Errorf("tag検索で取得できる求人にtagが含まれていません tag: %q, job.Tags: %q", tag, job.Tags))
			return
		}
	}

	if err := validate.StatusCode(resp2, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	// 2ページ目の件数が10件となっていること
	if len(result2.Jobs) != 10 {
		addError(ctx, step, ErrCritical, fmt.Errorf("tag検索一覧で2ページ目に取得できる求人の数が期待値(10)と異なります: %d", len(result1.Jobs)))
		return
	}

	// 3ページ目が存在しないことが示されていること
	if result2.HasNextPage {
		addError(ctx, step, ErrCritical, errors.New("次のページがないのにhas_next_pageの値がtrueになっています"))
		return
	}

	// ソート順が正しいこと
	err = validate.JobResultOrder(result2.Jobs)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索で取得できる求人の順序が正しくありません求人検索で取得できる求人の順序が正しくありません: %w", err))
		return
	}

	// 検索時に指定したタグを含んでいること
	for _, job := range result2.Jobs {
		if !strings.Contains(job.Tags, tag) {
			addError(ctx, step, ErrCritical, fmt.Errorf("tag検索で取得できる求人にtagが含まれていません tag: %q, job.Tags: %q", tag, job.Tags))
			return
		}
	}

	// 2ページ目と1ページ目が正しいソート順であること
	if result1.Jobs[0].UpdatedAt.Before(result2.Jobs[0].UpdatedAt) {
		addError(ctx, step, ErrCritical, fmt.Errorf("ページングにおけるソート順が正しくありません (result1.Jobs[0].UpdatedAt: %v, result2.Jobs[0].UpdatedAt: %v)", result1.Jobs[0].UpdatedAt, result2.Jobs[0].UpdatedAt))
		return
	}

}

// GET /cs/job_search 正常系 industry 検索
func verifyCSJobSearchIndustryOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// CLユーザーで求人を作成
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
	}

	cl, _, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成またはに失敗しました: %w", err))
		return
	}

	// CSユーザー
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
	}

	/* Act */
	query := api.JobSearchQuery{
		IndustryID: cl.Company.IndustryID,
	}
	resp, err := api.GetCSJobSearch(ctx, ag2, query)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	result := model.JobSearchResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索結果jsonの読み取りに失敗しました: %w", err))
		return
	}

	err = validate.JobResultOrder(result.Jobs)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人検索で取得できる求人の順序が正しくありません: %w", err))
		return
	}

	// 求人の内容を確認
	for _, job := range result.Jobs {
		if job.Company.Industry != cl.Company.Industry {
			addError(ctx, step, ErrCritical, fmt.Errorf("industry検索で取得できる求人のindustry_idが正しくありません expected(%s) != actual(%s)", cl.Company.Industry, job.Company.Industry))
			return
		}
	}
}

// POST /cs/application 正常系
func verifyCSApplicationOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// CLユーザーで求人を作成
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	// CSユーザー
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}
	_, err = action.CreateCSUserWithAuth(ctx, ag2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}

	// 求人応募
	resp, err := api.PostCSApplication(ctx, ag2, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人への応募に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
	type Application struct {
		ID int `json:"id"`
	}
	var application Application
	if err := json.NewDecoder(resp.Body).Decode(&application); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人応募APIのレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}
	if application.ID <= 0 {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人応募のIDが不正です: %d", application.ID))
		return
	}

}

// POST /cs/application 異常系 未ログインでの応募
func verifyCSApplicationNotLogin(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// CLユーザーで求人を作成
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	/* Act */
	// CSユーザー
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}
	resp, err := api.PostCSApplication(ctx, ag2, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人への応募に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

}

// POST /cs/application 異常系 存在しない求人ID
func verifyCSApplicationNotExistJobID(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, err = action.CreateCSUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.PostCSApplication(ctx, ag, 0) // job_id = 0 is not exist
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人への応募に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusNotFound); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

}

// POST /cs/application 異常系 jobがactiveでない
func verifyCSApplicationNotActiveJob(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// CLユーザーで求人を作成
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	// CSユーザー
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}
	_, err = action.CreateCSUserWithAuth(ctx, ag2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}

	// 求人を非アクティブにする
	is_active := false
	_, err = api.PatchCLJob(ctx, ag, job.ID, nil, nil, nil, nil, &is_active)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の更新に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.PostCSApplication(ctx, ag2, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人への応募に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnprocessableEntity); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// GET /cs/applications 正常系 複数企業の求人への応募一覧取得
func verifyCSApplicationsOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// 企業Aで求人作成
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, job1, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	// 企業Bで求人作成
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}
	_, job2, err := action.CreateJobWithNewUser(ctx, ag2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	// CSユーザーでログイン
	ag3, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}
	_, err = action.CreateCSUserWithAuth(ctx, ag3)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}

	// 求人に応募
	api.PostCSApplication(ctx, ag3, job1.ID)
	api.PostCSApplication(ctx, ag3, job2.ID)

	/* Act */
	resp, err := api.GetCSApplications(ctx, ag3, nil)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧の取得に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	result := model.ApplicationsResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧のレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
	if len(result.Applications) != 2 {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧で取得できる求人の数が期待値(2)と異なります: %d", len(result.Applications)))
		return
	}
	err = validate.JobForCS(result.Applications[0].Job, *job2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧で取得できる求人が正しくありません: %w", err))
		return
	}
	err = validate.JobForCS(result.Applications[1].Job, *job1)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧で取得できる求人が正しくありません: %w", err))
		return
	}

	err = validate.ApplicationResultOrder(result.Applications)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧で取得できる求人の順序が正しくありません: %w", err))
		return
	}
}

// GET /cs/applications 正常系 ページング確認
func verifyCSApplicationListPagingOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// 企業アカウントで求人作成
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to create agent: %w", err))
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	jobIDs := []int{}
	for i := 0; i < 30; i++ {
		job, err := action.CreateJob(ctx, ag)
		if err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
			return
		}
		jobIDs = append(jobIDs, job.ID)
	}

	// CSユーザー
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to create agent: %w", err))
		return
	}
	_, err = action.CreateCSUserWithAuth(ctx, ag2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CSアカウントの登録に失敗しました: %w", err))
		return
	}

	// 応募
	for _, jobID := range jobIDs {
		api.PostCSApplication(ctx, ag2, jobID)
	}

	/* Act */
	// 1ページ目(20件)
	res1, err := api.GetCSApplications(ctx, ag2, nil)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧の取得に失敗しました: %w", err))
		return
	}
	defer res1.Body.Close()

	// 2ページ目(10件)
	page := 1
	resp2, err := api.GetCSApplications(ctx, ag2, &page)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧の取得に失敗しました: %w", err))
		return
	}
	defer resp2.Body.Close()

	/* Assert */
	// 1ページ目
	if err := validate.StatusCode(res1, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	result1 := model.ApplicationsResponse{}
	if err := json.NewDecoder(res1.Body).Decode(&result1); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	if len(result1.Applications) != 20 {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧で1ページ目に取得できる求人の数が期待値(20)と異なります: %d", len(result1.Applications)))
		return
	}

	if !result1.HasNextPage {
		err := errors.New("応募数が21件以上にも関わらずhas_next_page が false になっています")
		addError(ctx, step, ErrCritical, err)
		return
	}

	err = validate.ApplicationResultOrder(result1.Applications)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧で取得できる求人の順序が正しくありません: %w", err))
		return
	}

	// ページ2
	if err := validate.StatusCode(resp2, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	result2 := model.ApplicationsResponse{}
	if err := json.NewDecoder(resp2.Body).Decode(&result2); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧APIのレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}

	if len(result2.Applications) != 10 {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧で2ページ目に取得できる求人の数が期待値(10)と異なります: %d", len(result2.Applications)))
		return
	}

	if result2.HasNextPage {
		addError(ctx, step, ErrCritical, errors.New("次のページがないのにhas_next_pageの値がtrueになっています"))
		return
	}

	err = validate.ApplicationResultOrder(result2.Applications)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧で取得できる求人の順序が正しくありません: %w", err))
		return
	}

	// 2ページ目と1ページ目が正しいソート順であること
	if result1.Applications[0].CreatedAt.Before(result2.Applications[0].CreatedAt) {
		addError(ctx, step, ErrCritical, errors.New("ページングにおけるソート順が正しくありません"))
		return
	}

}

// GET /cs/applications 異常系 未ログインでの応募一覧取得
func verifyCSApplicationsNotLogin(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		return
	}

	/* Act */
	resp, err := api.GetCSApplications(ctx, ag, nil)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("応募一覧の取得に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// POST /cl/company 正常系
func verifyCLCompanyOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	c := fixture.GenerateCompany()

	/* Act */
	resp, err := api.PostCLCompany(ctx, ag, c.Name, c.IndustryID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("企業の登録に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
	type Company struct {
		ID int `json:"id"`
	}
	var company Company
	if err := json.NewDecoder(resp.Body).Decode(&company); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("企業登録APIのレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}
	if company.ID <= 0 {
		addError(ctx, step, ErrCritical, fmt.Errorf("登録した企業のIDが不正です: %d", company.ID))
		return
	}
}

// POST /cl/signup 正常系
func verifyCLSignupOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	c, err := action.CreateCompany(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("企業の作成に失敗しました: %w", err))
		return
	}

	cl := fixture.GenerateCLUser()

	/* Act */
	resp, err := api.PostCLSignup(ctx, ag, cl.Email, cl.Password, cl.Name, c.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
	type CLUser struct {
		ID int `json:"id"`
	}
	var clUser CLUser
	if err := json.NewDecoder(resp.Body).Decode(&clUser); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録APIのレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}
	if clUser.ID <= 0 {
		addError(ctx, step, ErrCritical, fmt.Errorf("登録したCLアカウントのIDが不正です: %d", clUser.ID))
		return
	}

}

// POST /cl/signup 異常系 同一ユーザーでの登録
func verifyCLSignupDuplicateUser(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cl, err := action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.PostCLSignup(ctx, ag, cl.Email, cl.Password, cl.Name, cl.CompanyID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusConflict); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// POST /cl/signup 異常系 存在しない企業ID
func verifyCLSignupNotExistComapnyID(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cl := fixture.GenerateCLUser()

	/* Act */
	resp, err := api.PostCLSignup(ctx, ag, cl.Email, cl.Password, cl.Name, 0) // company_id = 0 is not exist
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusBadRequest); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// POST /cl/login 正常系
func verifyCLLoginOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cl, err := action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.PostCLLogin(ctx, ag, cl.Email, cl.Password)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントでのログインに失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// POST /cl/login 異常系 存在しないユーザー
func verifyCLLoginNotExist(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cl := fixture.GenerateCLUser()

	/* Act */
	resp, err := api.PostCLLogin(ctx, ag, cl.Email, cl.Password)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントでのログインに失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// POST /cl/login 異常系 誤ったパスワード
func verifyCLLoginWrongPassword(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	cl, err := action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.PostCLLogin(ctx, ag, cl.Email, "wrong"+cl.Password)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントでのログインに失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// POST /cl/logout 正常系
func verifyCLLogoutOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.PostCLLogout(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントでのログアウトに失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// POST /cl/logout 異常系 未ログインでのログアウト
func verifyCLLogoutNotLogin(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	/* Act */
	resp, err := api.PostCLLogout(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントでのログアウトに失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// POST /cl/job 正常系
func verifyCreateJobOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	job := fixture.GenerateJob()

	/* Act */
	resp, err := api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	if err := json.NewDecoder(resp.Body).Decode(job); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人作成APIのレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}
	if job.ID <= 0 {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人のIDが不正です: %d", job.ID))
		return
	}

	resp2, err := api.GetCLJob(ctx, ag, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to get job: %w", err))
		return
	}
	defer resp2.Body.Close()

	type JobResponse struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Salary      int    `json:"salary"`
		Tags        string `json:"tags"`
		IsActive    bool   `json:"is_active"`
	}
	var jobResponse JobResponse
	if err := json.NewDecoder(resp2.Body).Decode(&jobResponse); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to decode job response body: %w", err))
		return
	}

	if jobResponse.Title != job.Title {
		addError(ctx, step, ErrCritical, fmt.Errorf("title is not updated: jobResponse.Title: %s, job.Title: %s", jobResponse.Title, job.Title))
		return
	}
	if jobResponse.Description != job.Description {
		addError(ctx, step, ErrCritical, fmt.Errorf("description is not updated: jobResponse.Description: %s, job.Description: %s", jobResponse.Description, job.Description))
		return
	}
	if jobResponse.Salary != job.Salary {
		addError(ctx, step, ErrCritical, fmt.Errorf("salary is not updated: jobResponse.Salary: %d, job.Salary: %d", jobResponse.Salary, job.Salary))
		return
	}
	if jobResponse.Tags != job.Tags {
		addError(ctx, step, ErrCritical, fmt.Errorf("tags is not updated: jobResponse.Tags: %s, job.Tags: %s", jobResponse.Tags, job.Tags))
		return
	}
	if !jobResponse.IsActive {
		addError(ctx, step, ErrCritical, fmt.Errorf("is_active is not updated: %t", jobResponse.IsActive))
		return
	}
}

// POST /cl/job 異常系 未ログインでの求人作成
func verifyCreateJobNotLogin(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	job := fixture.GenerateJob()

	/* Act */
	resp, err := api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}
}

// PATCH /cl/job/:jobid 正常系
func verifyUpdateJobOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	/* Act */
	job.Title = "updated title"
	job.Description = "updated description"
	job.Salary = 200000
	job.Tags = "tag1,tag2"
	job.IsActive = false

	resp, err := api.PatchCLJob(ctx, ag, job.ID, &job.Title, &job.Description, &job.Salary, &job.Tags, &job.IsActive)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to update job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}
	resp2, err := api.GetCLJob(ctx, ag, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to get job: %w", err))
		return
	}
	defer resp2.Body.Close()

	type JobResponse struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Salary      int    `json:"salary"`
		Tags        string `json:"tags"`
		IsActive    bool   `json:"is_active"`
	}
	var jobResponse JobResponse
	if err := json.NewDecoder(resp2.Body).Decode(&jobResponse); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to decode job response body: %w", err))
		return
	}

	if jobResponse.Title != job.Title {
		addError(ctx, step, ErrCritical, fmt.Errorf("title is not updated: jobResponse.Title: %s, job.Title: %s", jobResponse.Title, job.Title))
		return
	}
	if jobResponse.Description != job.Description {
		addError(ctx, step, ErrCritical, fmt.Errorf("description is not updated: jobResponse.Description: %s, job.Description: %s", jobResponse.Description, job.Description))
		return
	}
	if jobResponse.Salary != job.Salary {
		addError(ctx, step, ErrCritical, fmt.Errorf("salary is not updated: jobResponse.Salary: %d, job.Salary: %d", jobResponse.Salary, job.Salary))
		return
	}
	if jobResponse.Tags != job.Tags {
		addError(ctx, step, ErrCritical, fmt.Errorf("tags is not updated: jobResponse.Tags: %q, job.Tags: %q", jobResponse.Tags, job.Tags))
		return
	}
	if jobResponse.IsActive != job.IsActive {
		addError(ctx, step, ErrCritical, fmt.Errorf("is_active is not updated: %t", jobResponse.IsActive))
		return
	}

}

// PATCH /cl/job/:jobid 異常系 未ログインでの求人更新
func verifyUpdateJobNotLogin(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	/* Act */
	job.Title = "updated title"
	job.Description = "updated description"
	job.Salary = 200000
	job.Tags = "tag1,tag2"
	job.IsActive = false

	resp, err := api.PatchCLJob(ctx, ag2, job.ID, &job.Title, &job.Description, &job.Salary, &job.Tags, &job.IsActive)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to update job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}
}

// PATCH /cl/job/:jobid 異常系 別企業のユーザーによる求人更新
func verifyUpdateJobNotOwnerUser(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// 企業A
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	// 企業B
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	job.Title = "updated title"
	job.Description = "updated description"
	job.Salary = 200000
	job.Tags = "tag1,tag2"
	job.IsActive = false

	resp, err := api.PatchCLJob(ctx, ag2, job.ID, &job.Title, &job.Description, &job.Salary, &job.Tags, &job.IsActive)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to update job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusForbidden); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}

}

// PATCH /cl/job/:jobid 異常系 存在しない求人ID
func verifyUpdateJobNotExistJobID(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	job := fixture.GenerateJob()
	job.Title = "updated title"
	job.Description = "updated description"
	job.Salary = 200000
	job.Tags = "tag1,tag2"
	job.IsActive = false

	resp, err := api.PatchCLJob(ctx, ag, 0, &job.Title, &job.Description, &job.Salary, &job.Tags, &job.IsActive) // job_id = 0 is not exist
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to update job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusNotFound); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}
}

// PATCH /cl/job/:jobid 異常系 アーカイブ済み求人の更新
func verifyUpdateJobArchivedJob(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// 企業A
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	job, err := action.CreateArchivedJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to archive job: %w", err))
		return
	}

	/* Act */
	job.Title = "updated title"
	job.Description = "updated description"
	job.Salary = 200000
	job.Tags = "tag1,tag2"
	job.IsActive = false

	resp, err := api.PatchCLJob(ctx, ag, job.ID, &job.Title, &job.Description, &job.Salary, &job.Tags, &job.IsActive)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to update job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnprocessableEntity); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}

}

// POST /cl/job/:jobid/archive 正常系
func verifyArchiveJobOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.POSTCLJobArchive(ctx, ag, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to archive job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}
}

// POST /cl/job/:jobid/archive 異常系 未ログインでの求人アーカイブ
func verifyArchiveJobNotLogin(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	/* Act */
	resp, err := api.POSTCLJobArchive(ctx, ag2, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to archive job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}
}

// POST /cl/job/:jobid/archive 異常系 別企業のユーザーによる求人アーカイブ
func verifyArchiveJobNotOwnerUser(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// 企業A
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	// 企業B
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.POSTCLJobArchive(ctx, ag2, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to archive job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusForbidden); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}
}

// POST /cl/job/:jobid/archive 異常系 存在しない求人ID
func verifyArchiveJobNotExistJobID(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.POSTCLJobArchive(ctx, ag, 0) // job_id = 0 is not exist
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to archive job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusNotFound); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}
}

// GET /cl/job/:jobid 正常系
func verifyGetJobOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}

	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}
	job.IsActive = true

	/* Act */
	resp, err := api.GetCLJob(ctx, ag, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to get job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}

	var jobResponse model.Job
	if err := json.NewDecoder(resp.Body).Decode(&jobResponse); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人取得結果jsonの読み取りに失敗しました: %w", err))
		return
	}

	err = validate.JobForCL(jobResponse, *job)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("作成したjobが正しく取得できません: %w", err))
		return
	}
}

// GET /cl/job/:jobid 正常系 アーカイブ済み求人でも取得可能
func verifyGetJobArchivedJob(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// 企業A
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}
	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}
	resp, err := api.POSTCLJobArchive(ctx, ag, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to archive job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Act */
	resp2, err := api.GetCLJob(ctx, ag, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to get job: %w", err))
		return
	}
	defer resp2.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp2, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}

}

// GET /cl/job/:jobid 異常系 未ログインでの求人取得
func verifyGetJobNotLogin(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}
	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	/* Act */
	resp, err := api.GetCLJob(ctx, ag2, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to get job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}

}

// GET /cl/job/:jobid 異常系 別企業のユーザーによる求人取得
func verifyGetJobNotOwnerUser(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// 企業A
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to new agent: %w", err))
		return
	}
	_, job, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	// 企業B
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.GetCLJob(ctx, ag2, job.ID)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to get job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusForbidden); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}

}

// GET /cl/job/:jobid 異常系 存在しない求人ID
func verifyGetJobNotExistJobID(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.GetCLJob(ctx, ag, 0) // job_id = 0 is not exist
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to get job: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusNotFound); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
		return
	}

}

// GET /cl/jobs 正常系
func verifyGetJobsOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	// ログインユーザー用agent
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}
	// ログインユーザーで1つ目の求人を作成
	cl1, job1, err := action.CreateJobWithNewUser(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}
	job1.IsActive = true

	// ログインユーザーで2つ目の求人を作成
	job2, err := action.CreateJob(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}
	job2.IsActive = true

	// 同一企業の別ユーザー用agent
	ag2, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}
	_, err = action.CreateCLUserForSpecifiedComapnyWithAuth(ctx, ag2, &cl1.Company)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to create cl user: %w", err))
		return
	}

	//  同一企業の別ユーザーで3つ目の求人を作成
	job3, err := action.CreateJob(ctx, ag2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
		return
	}

	// 別企業のユーザー用agent
	ag3, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	// 別企業のユーザーで4つ目の求人を作成
	// 別企業のユーザーによる求人なので取得されない
	_, _, err = action.CreateJobWithNewUser(ctx, ag3)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの作成または求人の作成に失敗しました: %w", err))
		return
	}

	/* Act */
	resp, err := api.GetCLJobs(ctx, ag, nil)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to get jobs: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
	}

	var jobsResponse model.JobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobsResponse); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧取得APIのレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}

	if len(jobsResponse.Jobs) != 3 {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧で取得できる求人の数が期待値(3)と異なります: %d", len(jobsResponse.Jobs)))
		return
	}
	if jobsResponse.Jobs[0].UpdatedAt.Before(jobsResponse.Jobs[1].UpdatedAt) {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧で取得できる求人の順序が正しくありません jobsResponse.Jobs[0].UpdatedAt: %v, jobsResponse.Jobs[1].UpdatedAt: %v", jobsResponse.Jobs[0].UpdatedAt, jobsResponse.Jobs[1].UpdatedAt))
		return
	}
	err = validate.JobForCS(jobsResponse.Jobs[0], *job3)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧で取得できる求人が正しくありません: %w", err))
		return
	}
	err = validate.JobForCS(jobsResponse.Jobs[1], *job2)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧で取得できる求人が正しくありません: %w", err))
		return
	}
	err = validate.JobForCS(jobsResponse.Jobs[2], *job1)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧で取得できる求人が正しくありません: %w", err))
		return
	}

}

// GET /cl/jobs 正常系 ページング確認
func verifyGetJobsPagingOK(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, fmt.Errorf("failed to create agent: %w", err))
		return
	}

	_, err = action.CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("CLアカウントの登録に失敗しました: %w", err))
		return
	}

	for i := 0; i < 60; i++ {
		_, err := action.CreateJob(ctx, ag)
		if err != nil {
			addError(ctx, step, ErrCritical, fmt.Errorf("求人の作成に失敗しました: %w", err))
			return
		}
	}

	/* Act */
	// 1ページ目(50件)
	res1, err := api.GetCLJobs(ctx, ag, nil)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧の取得に失敗しました: %w", err))
		return
	}
	defer res1.Body.Close()

	// 2ページ目(10件)
	page := 1
	resp2, err := api.GetCLJobs(ctx, ag, &page)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧の取得に失敗しました: %w", err))
		return
	}
	defer resp2.Body.Close()

	/* Assert */
	// 1ページ目
	if err := validate.StatusCode(res1, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	result1 := model.JobsResponse{}
	if err := json.NewDecoder(res1.Body).Decode(&result1); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧取得APIのレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}

	if len(result1.Jobs) != 50 {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧で1ページ目に取得できる求人の数が期待値(50)と異なります: %d", len(result1.Jobs)))
		return
	}

	if !result1.HasNextPage {
		err := errors.New("求人一覧で求人数が51件以上にも関わらずhas_next_page が false になっています")
		addError(ctx, step, ErrCritical, err)
		return
	}

	err = validate.JobResultOrder(result1.Jobs)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧で取得できる求人の順序が正しくありません: %w", err))
		return
	}

	// ページ2
	if err := validate.StatusCode(resp2, http.StatusOK); err != nil {
		addError(ctx, step, ErrCritical, err)
		return
	}

	result2 := model.JobsResponse{}
	if err := json.NewDecoder(resp2.Body).Decode(&result2); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧取得APIのレスポンスjsonの読み取りに失敗しました: %w", err))
		return
	}

	if len(result2.Jobs) != 10 {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧で2ページ目に取得できる求人の数が期待値(10)と異なります: %d", len(result2.Jobs)))
		return
	}

	if result2.HasNextPage {
		addError(ctx, step, ErrCritical, errors.New("次のページがないのにhas_next_pageの値がtrueになっています"))
		return
	}

	err = validate.JobResultOrder(result2.Jobs)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("求人一覧で取得できる求人の順序が正しくありません: %w", err))
		return
	}

	// 2ページ目と1ページ目が正しいソート順であること
	if result1.Jobs[0].CreatedAt.Before(result2.Jobs[0].CreatedAt) {
		addError(ctx, step, ErrCritical, fmt.Errorf("ページングにおけるソート順が正しくありません result1.Jobs[0].CreatedAt: %v, result2.Jobs[0].CreatedAt: %v", result1.Jobs[0].CreatedAt, result2.Jobs[0].CreatedAt))
		return
	}
}

// GET /cl/jobs 異常系 未ログインでの求人一覧取得
func verifyGetJobsNotLogin(ctx context.Context, step *isucandar.BenchmarkStep, opt Option) {
	/* Arrange */
	ag, err := NewAgent(opt.TargetHost, opt.PrepareRequestTimeout, opt.BenchID)
	if err != nil {
		addError(ctx, step, ErrServerCritical, err)
		return
	}

	/* Act */
	resp, err := api.GetCLJobs(ctx, ag, nil)
	if err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to get jobs: %w", err))
		return
	}
	defer resp.Body.Close()

	/* Assert */
	if err := validate.StatusCode(resp, http.StatusUnauthorized); err != nil {
		addError(ctx, step, ErrCritical, fmt.Errorf("failed to validate status code: %w", err))
	}

}
