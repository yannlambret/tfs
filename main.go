package main

import (
	"github.com/spf13/cobra"
	"github.com/yannlambret/tfs/pkg/tfs"

	cmd "github.com/yannlambret/tfs/cmd/tfs"
)

func main() {
	// Application configuration.
	cobra.OnInitialize(tfs.InitConfig)

	// CLI entry point.
	cmd.Execute()
}
