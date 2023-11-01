package tfs

import (
	"os"
	"time"

	"log/slog"

	"github.com/Masterminds/semver"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yannlambret/tfs/pkg/tfs"
)

var (
	quiet    bool
	LogLevel = new(slog.LevelVar)

	rootCmd = &cobra.Command{
		Use:           `tfs`,
		Short:         `Automatically fetch and configure the required version of the Terraform binary`,
		Example:       `cd <path> && tfs`,
		SilenceUsage:  true,
		SilenceErrors: true,

		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				slog.Error("This command supports one positional argument at most")
				return err
			}
			// Custom validation logic.
			if _, err := semver.NewVersion(args[0]); err != nil {
				slog.Error("Command argument should be a valid Terraform version")
				return err
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				v   *semver.Version
				err error
			)

			// Load local cache.
			if err := tfs.Cache.Load(); err != nil {
				return err
			}

			// The user wants to run a specific Terraform version.
			if len(args) != 0 {
				// Ignoring potential errors here because we have already
				// checked that the argument is a valid semantic version.
				v, _ = semver.NewVersion(args[0])
			} else {
				// Terraform configuratin parsing. If a specific version
				// is defined in the 'terraform' block, we will try to use it.
				v, err = tfs.GetTfVersion()
			}

			if err != nil {
				return err
			}

			if v != nil {
				// Create and activate a release for the target semantic version.
				release := tfs.NewRelease(v).Init()

				if err := release.Install(); err != nil {
					return err
				}
				if err := release.Activate(); err != nil {
					return err
				}
				// Remove extra releases in order to keep
				// a reasonable cache size.
				tfs.Cache.AutoClean()
			}

			return nil
		},
	}
)

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", true, "Reduce logging verbosity")
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	w := os.Stderr
	handler := tint.NewHandler(w, &tint.Options{
		Level:      LogLevel,
		NoColor:    !isatty.IsTerminal(w.Fd()),
		TimeFormat: time.Kitchen,
	})

	// Set global logger with custom options.
	slog.SetDefault(slog.New(handler))
}
