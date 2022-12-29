package tfs

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
	"github.com/yannlambret/tfs/pkg/tfs"
)

var (
	rootCmd = &cobra.Command{
		Use:           `tfs`,
		Short:         `Automatically fetch and configure the required version of Terraform binary`,
		Example:       `cd <path> && tfs`,
		SilenceUsage:  true,
		SilenceErrors: true,

		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				log.Error("This command supports one positional argument at most")
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
			var (
				v   *semver.Version
				err error
			)
			// The user wants to run a specific Terraform version
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
			}

			return nil
		},
	}
)

func Execute() {
	// Logging configuration.
	log.SetHandler(cli.New(os.Stdout))

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
