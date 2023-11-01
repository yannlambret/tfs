package tfs

import (
	"strings"

	"github.com/Masterminds/semver"
	"github.com/dustin/go-humanize"
	"github.com/spf13/viper"
)

// versionFromFileName extracts semantic version from Terraform binary name.
func versionFromFileName(fileName string) (*semver.Version, error) {
	return semver.NewVersion(strings.ReplaceAll(fileName, viper.GetString("terraform_file_name_prefix"), ""))
}

// formatSize returns size in a human readable format.
func formatSize(size uint64) string {
	return humanize.Bytes(size)
}
