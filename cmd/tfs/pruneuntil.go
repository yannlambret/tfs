package tfs

import (
	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/yannlambret/tfs/pkg/tfs"
)

var pruneuntilCmd = &cobra.Command{
	Use:     "prune-until",
	Short:   "Remove Terraform binary for the specified version and all the previous ones",
	Example: `prune-until 1.3.0`,

	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			log.Error("This command supports one positional argument exactly")
			return err
		}
		// Custom validation logic.
		if _, err := semver.NewVersion(args[0]); err != nil {
			log.Error("Command argument should be a valid Terraform version")
			return err
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		// Ignoring potential errors here because we have already
		// checked that the argument is a valid semantic version.
		v, _ := semver.NewVersion(args[0])
		return tfs.Cache.PruneUntil(v)
	},
}

func init() {
	rootCmd.AddCommand(pruneuntilCmd)
}
