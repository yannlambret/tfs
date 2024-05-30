package tfs

import (
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"
)

// versionFromFileName extracts semantic version from Terraform binary name.
func versionFromFileName(fileName string) (*version.Version, error) {
	return version.NewVersion(strings.ReplaceAll(fileName, viper.GetString("terraform_file_name_prefix"), ""))
}

// formatSize returns size in a human readable format.
func formatSize(size uint64) string {
	return humanize.Bytes(size)
}
