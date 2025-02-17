package benchmarker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"risuwork-benchmarker/internal"
	"risuwork-benchmarker/internal/logger"
	"risuwork-benchmarker/scenario"
	"risuwork-benchmarker/scenario/api"
	"strings"
	"syscall"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
)

func RunContext(ctx context.Context, opt Option) error {
	ctx, cancel := internal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// 現在の設定を大会運営向けロガーに出力
	logger.Admin().Debug(opt.String())

	bOpts := make([]isucandar.BenchmarkOption, 0, 2)

	// isucandar.Benchmark はステップ内の panic を自動で recover する機能があるが、今回は利用しない
	bOpts = append(bOpts, isucandar.WithoutPanicRecover())

	if opt.PrepareOnly {
		opt.LoadTimeout = 1 * time.Millisecond // isucandarは必ず負荷試験を実施してしまうので、タイムアウトを小さく設定する
	} else if opt.LoadTimeout <= 0 {
		opt.LoadTimeout = 1 * time.Minute // 設定がない場合はデフォルトの1分間に設定
	}
	bOpts = append(bOpts, isucandar.WithLoadTimeout(opt.LoadTimeout))

	// ベンチマークの生成
	benchmark, err := isucandar.NewBenchmark(bOpts...)
	if err != nil {
		logger.Admin().Error(err.Error())
		return err
	}

	benchmark.IgnoreErrorCode(scenario.ErrIgnore)
	benchmark.OnError(func(err error, step *isucandar.BenchmarkStep) {
		if failure.IsCode(err, scenario.ErrIgnore) || failure.IsCode(err, failure.TemporaryErrorCode) || failure.IsCode(err, failure.CanceledErrorCode) {
			// no-op ErrIgnore, TemporaryErrorCode, CanceledErrorCodeはエラーとして扱わず、ログなどにも出力しない
			// TimeoutErrorCodeはログに出力するため、ここでは対象外
			return
		}

		// timeoutは負荷走行中の場合のみログ出力だけして減点や終了は無し
		if failure.IsCode(err, failure.TimeoutErrorCode) && failure.IsCode(err, isucandar.ErrLoad) {
			var uErr *url.Error
			if errors.As(err, &uErr) {
				if parse, err := url.Parse(uErr.URL); err == nil {
					path := parse.Path
					// パスパラメータが個別にログに表出しないように調整
					if strings.HasPrefix(path, "/api/cl/job/") {
						if strings.HasSuffix(path, "/archive") {
							logger.Player().Warn(fmt.Sprintf("Timeout: %s /api/cl/job/:jobid/archive", uErr.Op))
						} else {
							logger.Player().Warn(fmt.Sprintf("Timeout: %s /api/cl/job/:jobid", uErr.Op))
						}
					} else {
						logger.Player().Warn(fmt.Sprintf("Timeout: %s %s", uErr.Op, parse.Path))
					}
					return
				}
			}
			logger.Player().Warn("Timeout") // *url.ErrorじゃないかURLのパースでエラーになった場合も情報は少ないがユーザーに伝えるためにログ出力
			return
		}

		if failure.IsCode(err, scenario.ErrServerCritical) {
			step.Cancel()
			logger.Player().Error("benchmarker内部でエラーが発生しました。運営にお問い合わせください。")
			logger.Admin().Error(fmt.Sprintf("benchmarker内部でエラーが発生しました: %+v", err)) // スタックトレース付き
			return
		}

		// Note: Prepare時のタイムアウトはErrCriticalも兼ねているのでクリティカル扱いでここで判定
		if failure.IsCode(err, scenario.ErrCritical) {
			step.Cancel()
			logger.Player().Error(fmt.Sprintf("Critical: %s", GetFailureErrorMessage(err)))
			logger.Admin().Error(fmt.Sprintf("%+v", err))
			return
		}
	})

	// Scenarioを追加
	sc := &scenario.Scenario{Option: opt.Option}
	benchmark.Prepare(sc.Prepare)
	if !opt.PrepareOnly {
		benchmark.Load(sc.Load)
	}
	benchmark.Validation(sc.Validation)

	// ベンチマーク開始
	result := benchmark.Start(ctx)

	{ // finalize処理
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		select {
		case <-ticker.C: // 3秒待ってから終了処理を行う
			ag := scenario.MustNewAgent(opt.TargetHost, opt.InitializeRequestTimeout, opt.BenchID)
			if _, err = api.PostFinalize(ctx, ag); err != nil {
				err = fmt.Errorf("POST /finalize: %w", err)
				logger.Player().Warn(err.Error())
			}
		case <-ctx.Done():
		}
	}

	select {
	case <-ctx.Done():
		if cause := context.Cause(ctx); cause != nil {
			logger.Player().Warn("途中で終了しています", "cause", cause)
		}
	default:
	}

	{ // 結果表示
		// PrepareOnly用に結果表示
		if opt.PrepareOnly {
			if result.Errors.Count()[scenario.ErrCritical.ErrorCode()] > 0 ||
				result.Errors.Count()[scenario.ErrServerCritical.ErrorCode()] > 0 {
				return errors.New("チェックに失敗しました")
			}
			logger.Player().Info("チェックに成功しました")
			return nil
		}

		// 大会運営向けに全てのスコアを表示
		for tag, count := range result.Score.Breakdown() {
			logger.Player().Info(fmt.Sprintf("%s: %d", tag, count))
		}
		logger.Player().Info(fmt.Sprintf("timeout: %d", result.Errors.Count()[failure.TimeoutErrorCode.ErrorCode()]))

		// 選手向けにスコアを表示
		score := sumScore(result)
		logger.Player().Info("ベンチマーカー終了", slog.Int64("score", score))

		// 0点以下(fail)ならエラーで終了
		if opt.ExitErrorOnFail && score <= 0 {
			return errors.New("fail: score is 0 or less")
		}
	}

	return nil
}

func sumScore(result *isucandar.BenchmarkResult) int64 {
	score := result.Score
	// 各タグに倍率を設定
	score.Set(scenario.ScoreAuth, 1)
	score.Set(scenario.ScoreNormalGET, 3)
	score.Set(scenario.ScoreNormalPost, 3)
	score.Set(scenario.ScoreNormalPatch, 3)
	score.Set(scenario.ScoreJobApplication, 5)

	// 加点分の合算
	addition := score.Sum()

	// エラーによる減点は無し
	deduction := 0

	if result.Errors.Count()[scenario.ErrCritical.ErrorCode()] > 0 ||
		result.Errors.Count()[scenario.ErrServerCritical.ErrorCode()] > 0 {
		// クリティカルエラーがある場合は0点
		return 0
	}

	// 合計(0を下回ったら0点にする)
	return max(addition-int64(deduction), 0)
}

func GetFailureErrorMessage(err error) string {
	e := err
	for {
		var fErr *failure.Error
		if errors.As(e, &fErr) {
			e = fErr.Unwrap()
		} else {
			break
		}
	}
	return e.Error()
}
