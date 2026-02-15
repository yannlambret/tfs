package tfs

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const testFilePrefix = "terraform_"

// initTestFS sets up a temporary test filesystem using Afero.
func initTestFS(tb testing.TB) (string, func()) {
	tb.Helper()
	tempDir := tb.TempDir()

	// Create a real OS-based filesystem using a temp directory.
	AppFs = &AferoFs{Fs: afero.NewBasePathFs(afero.NewOsFs(), tempDir)}

	// Set required Viper config values.
	viper.Set("terraform_file_name_prefix", testFilePrefix)
	viper.Set("user_bin_directory", filepath.Join(tempDir, "bin"))

	// Ensure directories exist.
	if err := AppFs.MkdirAll(viper.GetString("user_bin_directory"), 0755); err != nil {
		tb.Fatalf("Failed to create bin dir: %v", err)
	}

	return tempDir, func() {
		viper.Reset()
	}
}

func writeTestFile(tb testing.TB, path string, content []byte) {
	tb.Helper()
	if err := afero.WriteFile(AppFs, path, content, 0644); err != nil {
		tb.Fatalf("Failed to write file: %v", err)
	}
}
