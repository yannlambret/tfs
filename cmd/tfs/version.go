package tfs

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewVersionCommand returns a new cobra.Command for the "version" subcommand.
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use: "version",

		Run: func(cmd *cobra.Command, args []string) {
			// Print current tfs version on standard output.
			slog.Info("tfs "+viper.GetString("version"), "version", viper.GetString("version"))
		},
	}
}
