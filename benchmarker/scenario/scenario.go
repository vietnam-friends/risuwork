package scenario

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"risuwork-benchmarker/internal/logger"
	"risuwork-benchmarker/scenario/model"
	"risuwork-benchmarker/scenario/validate"
	wx "risuwork-benchmarker/scenario/workerx"
	"sync"
	"syscall"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
)

// シナリオレベルで発生するエラーコードの定義
const (
	ErrCritical       failure.StringCode = "critical"
	ErrServerCritical failure.StringCode = "server-critical" // benchmarkerのエラー. これが発生するとbenchmarker実装がおかしい
	ErrIgnore         failure.StringCode = "ignore"
)

// シナリオで発生するスコアのタグ
const (
	ScoreNormalGET      score.ScoreTag = "GET success"
	ScoreNormalPost     score.ScoreTag = "POST success"
	ScoreNormalPatch    score.ScoreTag = "PATCH success"
	ScoreAuth           score.ScoreTag = "Authentication success"
	ScoreJobApplication score.ScoreTag = "Job Application success"
)

const (
	CreateCompanyMinClients     = 0
	CreateCompanyMaxClients     = 2
	CreateJobMinJobs            = 1
	CreateJobMaxJobs            = 5
	ApplyJobsNewCustomerRate    = 0.25
	ApplyJobsMinSearchApplyLoop = 5
	ApplyJobsMaxSearchApplyLoop = 10
)

type Scenario struct {
	Option       Option
	Companies    Container[model.Company]
	Clients      Container[model.CLUser]
	Jobs         Container[model.Job]
	Customers    Container[model.CSUser]
	Applications Container[model.Application]
}

type Option struct {
	BenchID                  string
	TargetHost               string
	RequestTimeout           time.Duration
	PrepareRequestTimeout    time.Duration
	InitializeRequestTimeout time.Duration
}

func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	// POST /initialize へのリクエストを実行
	logger.Player().Info("初期化リクエストを実行します")
	if err := s.postInitialize(ctx, step); err != nil {
		step.Cancel()
		logger.Player().Error("初期化リクエストに失敗しました")
		return err
	}
	logger.Player().Info("初期化リクエストに成功しました")

	// 負荷試験前の整合性チェックを実行
	logger.Player().Info("負荷走行前の整合性チェックを実行します")
	verify(ctx, step, s.Option)
	sleepWithContext(ctx, 1*time.Second)
	if HasCriticalError(step) {
		logger.Player().Error("負荷走行前の整合性チェックに失敗しました")
		return fmt.Errorf("負荷走行前の整合性チェックに失敗しました")
	}
	logger.Player().Info("負荷走行前の整合性チェックに成功しました")
	return nil
}

func (s *Scenario) Load(ctx context.Context, step *isucandar.BenchmarkStep) error {
	workers := []*wx.Worker{
		wx.MustNewWorker(s.clCreateCompanyScenario(step), wx.WithFixedDelay(250*time.Millisecond), wx.WithExponentialParallelismUpdate(2)),
		//wx.MustNewWorker(s.clSignupScenario(step), wx.WithFixedDelay(250*time.Millisecond), wx.WithExponentialParallelismUpdate(2)),
		wx.MustNewWorker(s.clCreateJobScenario(step), wx.WithFixedDelay(100*time.Millisecond), wx.WithExponentialParallelismUpdate(2)),
		wx.MustNewWorker(s.csApplyJobsScenario(step), wx.WithFixedDelay(100*time.Millisecond), wx.WithExponentialParallelismUpdate(2)),
		wx.MustNewWorker(s.clModifyJobScenario(step), wx.WithFixedDelay(100*time.Millisecond), wx.WithExponentialParallelismUpdate(2)),
	}

	logger.Player().Info("負荷走行を実行します。負荷レベル: 1", "level", 1)
	wg := &sync.WaitGroup{}
	for _, w := range workers {
		wg.Add(1)
		w := w // FIXME: go1.22 で修正される
		go func() {
			defer wg.Done()
			w.Process(ctx)
		}()
	}

	// 5秒ごとに負荷レベルを変更
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		lv := 1 // 初期負荷レベルは1
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if step.Result().Errors.Count()[failure.TimeoutErrorCode.ErrorCode()] > 0 {
					if lv > 1 {
						lv--
						logger.Player().Info(fmt.Sprintf("エラーが発生しているため、負荷レベルを下げます。負荷レベル: %d。以降負荷レベルは変動しません", lv))
						for _, w := range workers {
							from, to := w.UpdateParallelism(lv)
							logger.Admin().Debug(fmt.Sprintf("parallelism updated: %d -> %d", from, to), slog.Int64("from", int64(from)), slog.Int64("to", int64(to)))
						}
					} else {
						logger.Player().Info("エラーが発生しているため以降負荷レベルを変更しません")
					}
					return // tickerを終了
				}
				lv++
				logger.Player().Info(fmt.Sprintf("負荷レベルを変更します。負荷レベル: %d", lv), "level", lv)
				for _, w := range workers {
					from, to := w.UpdateParallelism(lv)
					logger.Admin().Debug(fmt.Sprintf("parallelism updated: %d -> %d", from, to), slog.Int64("from", int64(from)), slog.Int64("to", int64(to)))
				}
			}
		}
	}(ctx)

	wg.Wait()

	if HasCriticalError(step) {
		logger.Player().Error("負荷走行中にクリティカルなエラーが発生しました")
		return fmt.Errorf("負荷走行中にクリティカルなエラーが発生しました")
	}
	logger.Player().Info("負荷走行に成功しました")
	return nil
}

func (s *Scenario) Validation(_ context.Context, _ *isucandar.BenchmarkStep) error {
	return nil
}

func addScore(ctx context.Context, step *isucandar.BenchmarkStep, tag score.ScoreTag) {
	select {
	case <-ctx.Done():
		return
	default:
		step.AddScore(tag)
	}
}

func addError(ctx context.Context, step *isucandar.BenchmarkStep, code failure.Code, err error) {
	select {
	case <-ctx.Done():
		return
	default:
		var nerr net.Error
		var uerr *url.Error
		var opErr *net.OpError
		var syscallErr *os.SyscallError
		if errors.As(err, &nerr) && (nerr.Timeout() || nerr.Temporary()) {
			step.AddError(err)
		} else if errors.As(err, &uerr) && errors.Is(uerr.Unwrap(), io.EOF) {
			step.AddError(failure.NewError(failure.CanceledErrorCode, err))
		} else if errors.As(err, &opErr) && errors.As(opErr.Unwrap(), &syscallErr) && errors.Is(syscallErr.Unwrap(), syscall.ECONNRESET) {
			step.AddError(failure.NewError(failure.CanceledErrorCode, err))
		} else if errors.Is(err, context.Canceled) {
			step.AddError(err)
		} else if failure.IsCode(err, ErrIgnore) {
			step.AddError(err)
		} else {
			step.AddError(failure.NewError(code, err))
		}
	}
}

func HasCriticalError(step *isucandar.BenchmarkStep) bool {
	return step.Result().Errors.Count()[ErrCritical.ErrorCode()] > 0 ||
		step.Result().Errors.Count()[ErrServerCritical.ErrorCode()] > 0
}

type CreateResponseBody struct {
	Message string `json:"message"`
	ID      int    `json:"id"`
}

func check(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		var sErr validate.StatusCodeUnMatchError
		if errors.As(err, &sErr) && sErr.Actual == http.StatusUnprocessableEntity {
			return failure.NewError(ErrIgnore, err) // 422は処理中断するがIgnore
		}
		return err
	}
	return nil
}

func checkWithModel[T any](resp *http.Response, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := validate.StatusCode(resp, http.StatusOK); err != nil {
		var sErr validate.StatusCodeUnMatchError
		if errors.As(err, &sErr) && sErr.Actual == http.StatusUnprocessableEntity {
			return nil, failure.NewError(ErrIgnore, err) // 422は処理中断するがIgnore
		}
		return nil, err
	}
	var t T
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func sleepWithContext(ctx context.Context, duration time.Duration) {
	t := time.NewTimer(duration) // FIXME: Go1.23以降はtime.Afterで良さそう
	defer t.Stop()

	select {
	case <-t.C:
		break
	case <-ctx.Done():
		return
	}
}
