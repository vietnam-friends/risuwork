package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RootCmd(v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "bench",
		Short:             "Benchmark",
		Args:              cobra.NoArgs,
		SilenceUsage:      true, // don't show help content when error occurred
		SilenceErrors:     true, // Print error by own slog logger
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	}
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})

	cmd.AddCommand(RunCmd(v))
	cmd.AddCommand(SuperviseCmd(v))

	return cmd
}
