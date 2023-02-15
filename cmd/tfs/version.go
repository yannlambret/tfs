package tfs

import (
	"github.com/apex/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var versionCmd = &cobra.Command{
	Use: "version",

	Run: func(cmd *cobra.Command, args []string) {
		// Print current tfs version on standard output.
		log.Info("tfs " + viper.GetString("version"))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
