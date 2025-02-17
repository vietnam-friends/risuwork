package main

import (
	"risuwork-benchmarker"
	"risuwork-benchmarker/scenario"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// 各種オプションのデフォルト値
const (
	DefaultRequestTimeout           = 3 * time.Second
	DefaultPrepareRequestTimeout    = 10 * time.Second
	DefaultInitializeRequestTimeout = 30 * time.Second
	DefaultExitErrorOnFail          = true
	DefaultPrepareOnly              = false
	DefaultLoadInterval             = 1 * time.Minute
	DefaultBenchID                  = "-"
)

func RunCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <targetHost> [flags]",
		Short: "Run benchmarking",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetHost := args[0]
			requestTimeout := v.GetDuration("request-timeout")
			prepareRequestTimeout := v.GetDuration("prepare-request-timeout")
			initializeRequestTimeout := v.GetDuration("initialize-request-timeout")
			exitErrorOnFail := v.GetBool("exit-error-on-fail")
			prepareOnly := v.GetBool("prepare-only")
			loadTimeout := v.GetDuration("load-interval")
			benchID := v.GetString("bench-id")

			return benchmarker.RunContext(cmd.Context(), benchmarker.Option{
				Option: scenario.Option{
					TargetHost:               targetHost,
					RequestTimeout:           requestTimeout,
					PrepareRequestTimeout:    prepareRequestTimeout,
					InitializeRequestTimeout: initializeRequestTimeout,
					BenchID:                  benchID,
				},
				ExitErrorOnFail: exitErrorOnFail,
				PrepareOnly:     prepareOnly,
				LoadTimeout:     loadTimeout,
			})
		},
	}

	cmd.Flags().Duration("request-timeout", DefaultRequestTimeout, "Default request timeout")
	cmd.Flags().Duration("prepare-request-timeout", DefaultPrepareRequestTimeout, "Default request timeout")
	cmd.Flags().Duration("initialize-request-timeout", DefaultInitializeRequestTimeout, "Initialize request timeout")
	cmd.Flags().Bool("exit-error-on-fail", DefaultExitErrorOnFail, "Exit with error if benchmark fails")
	cmd.Flags().Bool("prepare-only", DefaultPrepareOnly, "Prepare step only if true")
	cmd.Flags().Duration("load-interval", DefaultLoadInterval, "Load interval")
	cmd.Flags().String("bench-id", DefaultBenchID, "Bench ID")
	_ = v.BindPFlags(cmd.Flags())

	return cmd
}
