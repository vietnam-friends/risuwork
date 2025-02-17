package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"risuwork-benchmarker/internal/logger"
	"strings"

	"github.com/PumpkinSeed/slog-context"
	"github.com/go-errors/errors"
	"github.com/isucon/isucandar/failure"
	"github.com/spf13/viper"
)

func init() {
	failure.BacktraceCleaner.Add(failure.SkipGOROOT)

	logger.SetAdmin(slog.New(slogcontext.NewHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))).With("for", "admin"))
	logger.SetPlayer(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})).With("for", "player"))
	slog.SetDefault(logger.Admin())
}

func main() {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	rootCmd := RootCmd(v)
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		logger.Player().Error(fmt.Sprintf("%s", err))
		logger.Admin().Error(fmt.Sprintf("%+v", err))
		var goErr *errors.Error
		if errors.As(err, &goErr) {
			logger.Admin().Error("stack trace", slog.String("stack", goErr.ErrorStack()))
		}
		os.Exit(1)
	}
}
