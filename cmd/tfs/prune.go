
package tfs

import (
	"github.com/spf13/cobra"
	"github.com/yannlambret/tfs/pkg/tfs"
)

// NewPruneCommand returns a new cobra.Command for the "prune" subcommand.
// It receives the cache instance that will be used by the command.
func NewPruneCommand(cache *tfs.LocalCache) *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Remove all Terraform binaries from the local cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load local cache.
			if err := cache.Load(); err != nil {
				return err
			}
			return cache.Prune()
		},
	}
}
