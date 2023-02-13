package tfs

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/spf13/viper"
)

// versionFromFileName extracts semantic version from Terraform binary file name.
func versionFromFileName(fileName string) (*semver.Version, error) {
	return semver.NewVersion(strings.ReplaceAll(fileName, viper.GetString("terraform_file_name_prefix"), ""))
}

// formatSize returns size in a human readable format.
func formatSize(size float64) string {
	var (
		k float64 = 1024
		m float64 = k * 1024
		g float64 = m * 1024
	)

	switch true {
	case (size >= m && size < g):
		return fmt.Sprintf("%.0fM", size/m)
	case (size >= g):
		return fmt.Sprintf("%.1fG", size/g)
	default:
		return fmt.Sprintf("%.0fB", size)
	}
}
