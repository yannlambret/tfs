package tfs

import (
	"github.com/spf13/cobra"
	"github.com/yannlambret/tfs/pkg/tfs"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List Terraform cached versions",
	Aliases: []string{"ls"},

	RunE: func(cmd *cobra.Command, args []string) error {
		return tfs.Cache.List()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
