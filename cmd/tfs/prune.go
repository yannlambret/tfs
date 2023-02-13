package tfs

import (
	"github.com/spf13/cobra"
	"github.com/yannlambret/tfs/pkg/tfs"
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove all Terraform binaries from the local cache",

	RunE: func(cmd *cobra.Command, args []string) error {
		// Load local cache.
		if err := tfs.Cache.Load(); err != nil {
			return err
		}

		return tfs.Cache.Prune()
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)
}
