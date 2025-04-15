package tfs

import (
	"log/slog"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/yannlambret/tfs/pkg/tfs"
)

// NewPruneUntilCommand returns a new cobra.Command for the "prune-until" subcommand.
// It receives the cache instance that will be used by the command.
func NewPruneUntilCommand(cache *tfs.LocalCache) *cobra.Command {
	return &cobra.Command{
		Use:     "prune-until",
		Short:   "Remove all Terraform binary versions prior to the one specified",
		Example: "prune-until 1.5.0",

		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				slog.Error("This command supports one positional argument exactly")
				return err
			}
			// Custom validation logic.
			if _, err := version.NewVersion(args[0]); err != nil {
				slog.Error("Command argument should be a valid Terraform version")
				return err
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			// Ignoring potential errors here because we have already
			// checked that the argument is a valid semantic version.
			v, _ := version.NewVersion(args[0])

			// Load local cache.
			if err := cache.Load(); err != nil {
				return err
			}

			return cache.PruneUntil(v)
		},
	}
}
