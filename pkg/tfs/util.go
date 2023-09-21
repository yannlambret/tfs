package tfs

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/dustin/go-humanize"
	"github.com/spf13/viper"
)

// versionFromFileName extracts semantic version from Terraform binary name.
func versionFromFileName(fileName string) (*semver.Version, error) {
	return semver.NewVersion(strings.ReplaceAll(fileName, viper.GetString("terraform_file_name_prefix"), ""))
}

// Align is use to give a consistent format to messages
// printed by the application.
func Align(padding int, message string) string {
	return fmt.Sprintf("%-*s", padding, message)
}

// formatSize returns size in a human readable format.
func formatSize(size uint64) string {
	return humanize.Bytes(size)
}
