package tfs

import (
	"os"
	"time"

	"log/slog"

	"github.com/hashicorp/go-version"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yannlambret/tfs/pkg/tfs"
)

var (
	quiet bool

	rootCmd = &cobra.Command{
		Use:           "tfs",
		Short:         "Automatically fetch and configure the required version of the Terraform binary",
		Example:       "cd <path> && tfs",
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
			if _, err := version.NewVersion(args[0]); err != nil {
				slog.Error("Command argument should be a valid Terraform version")
				return err
			}
			return nil
		},
	}
)

func Execute() {
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", true, "Reduce logging verbosity")
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))

	// Make sure configuration is initialized.
	tfs.InitConfig()

	// Retrieve the cache directory from configuration.
	cacheDir := viper.GetString("cache_directory")

	// Create a new cache instance.
	cache := tfs.NewLocalCache(cacheDir)

	// Add subcommands, injecting the cache instance when required.
	rootCmd.AddCommand(NewListCommand(cache))
	rootCmd.AddCommand(NewPruneCommand(cache))
	rootCmd.AddCommand(NewPruneUntilCommand(cache))
	rootCmd.AddCommand(NewVersionCommand())

	// Set the root command’s RunE function to use the cache.
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		var (
			v   *version.Version
			err error
		)

		// Load the cache contents.
		if err := cache.Load(); err != nil {
			return err
		}

		// Determine the target Terraform version.
		if len(args) != 0 {
			// We already validated that the argument is a valid semantic version.
			v, _ = version.NewVersion(args[0])
		} else {
			// If no argument is provided, try to get the version from configuration.
			if v, err = cache.GetTfVersion(); err != nil {
				return err
			}
		}

		if v != nil {
			// Create a new release in the cache.
			release := cache.NewRelease(v)

			if err := release.Install(); err != nil {
				return err
			}
			if err := release.Activate(); err != nil {
				return err
			}
			// Clean up extra releases.
			cache.AutoClean()
		} else {
			slog.Info("Did not find any version constraint in Terraform configuration")
			if !cache.IsEmpty() {
				// Use the most recent version of Terraform.
				if err := cache.LastRelease.Activate(); err != nil {
					return err
				}
			} else {
				slog.Info("Did not find any Terraform binary")
			}
		}

		return nil
	}

	// Execute the command.
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(tfs.InitConfig)

	// Logger initialization.
	var handler slog.Handler

	LogLevel := new(slog.LevelVar)
	w := os.Stderr

	if isatty.IsTerminal(w.Fd()) {
		handler = tint.NewHandler(w, &tint.Options{
			Level:      LogLevel,
			NoColor:    false,
			TimeFormat: time.Kitchen,
		})
	} else {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:     LogLevel,
			AddSource: false,
		})
	}

	// Set the global logger with custom options.
	slog.SetDefault(slog.New(handler))
}
