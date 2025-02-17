package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"risuwork-benchmarker"
	"risuwork-benchmarker/internal/logger"
	"risuwork-benchmarker/scenario"
	"time"

	"github.com/isucon/isucandar/failure"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	failure.BacktraceCleaner.Add(failure.SkipGOROOT)
	// Show only player logs
	logger.SetAdmin(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	logger.SetPlayer(slog.New(tint.NewHandler(os.Stdout, &tint.Options{TimeFormat: time.Kitchen, NoColor: !isatty.IsTerminal(os.Stdout.Fd())})))
	slog.SetDefault(logger.Admin())
}

func main() {
	rootCmd := RootCmd()
	rootCmd.SetOut(colorable.NewColorableStdout())
	rootCmd.SetErr(colorable.NewColorableStderr())
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		logger.Player().Error(err.Error())
		os.Exit(1)
	}
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "lightbench <target_endpoint>",
		Short:         "lightbench is lightweight benchmark for local development",
		Long:          `lightbenchは手元でのローカル開発やCI上で簡易的なベンチマーカーとして利用できます。`,
		Example:       "  ./lightbench localhost:8080",
		SilenceUsage:  true, // don't show help content when error occurred
		SilenceErrors: true, // Print error by own slog logger
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			shortLoad := viper.GetBool("short-load")
			longLoad := viper.GetBool("long-load")
			isDebug := viper.GetBool("debug")
			var loadTimeout time.Duration
			if shortLoad {
				loadTimeout = 10 * time.Second
			} else if longLoad {
				loadTimeout = 1 * time.Minute
			}

			if isDebug {
				logger.SetAdmin(slog.New(tint.NewHandler(os.Stdout, &tint.Options{TimeFormat: time.Kitchen, NoColor: !isatty.IsTerminal(os.Stdout.Fd())})))
				slog.SetDefault(logger.Admin())
			}
			opt := benchmarker.Option{
				Option: scenario.Option{
					TargetHost:               args[0],
					RequestTimeout:           3 * time.Second,
					PrepareRequestTimeout:    10 * time.Second,
					InitializeRequestTimeout: 30 * time.Second,
					BenchID:                  "lightbench",
				},
				PrepareOnly:     !shortLoad && !longLoad,
				ExitErrorOnFail: true,
				LoadTimeout:     loadTimeout,
			}
			return benchmarker.RunContext(cmd.Context(), opt)
		},
	}
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	cmd.Flags().Bool("debug", false, "デバッグログを表示します")
	cmd.Flags().BoolP("short-load", "s", false, "短期間(10s)負荷をかけて擬似的な点数を計算するモード。出力される点数はあくまで参考値です")
	cmd.Flags().BoolP("long-load", "l", false, "長期間(60s)負荷をかけて擬似的な点数を計算するモード。出力される点数はあくまで参考値です")
	cmd.MarkFlagsMutuallyExclusive("short-load", "long-load")
	_ = viper.BindPFlags(cmd.Flags())

	return cmd
}
