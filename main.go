package main

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/cobra"
	"github.com/yannlambret/tfs/pkg/tfs"

	cmd "github.com/yannlambret/tfs/cmd/tfs"
)

func main() {
	// Logging configuration.
	log.SetHandler(cli.New(os.Stdout))

	// Application configuration.
	cobra.OnInitialize(tfs.InitConfig)

	// CLI entry point.
	cmd.Execute()
}
