package tfs

import (
	"github.com/spf13/cobra"
	"github.com/yannlambret/tfs/pkg/tfs"
)

// NewListCommand returns a new cobra.Command for the "list" subcommand.
// It receives the cache instance that will be used by the command.
func NewListCommand(cache *tfs.LocalCache) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List Terraform cached versions",
		Aliases: []string{"ls"},

		RunE: func(cmd *cobra.Command, args []string) error {
			// Load local cache.
			if err := cache.Load(); err != nil {
				return err
			}

			return cache.List()
		},
	}
}
