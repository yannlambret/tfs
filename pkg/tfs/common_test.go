package tfs

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func getFilePrefix() string {
	return viper.GetString("terraform_file_name_prefix")
}

func writeTestFile(tb testing.TB, path string, content []byte) {
	tb.Helper()
	if err := afero.WriteFile(AppFs, path, content, 0644); err != nil {
		tb.Fatalf("Failed to write file: %v", err)
	}
}
