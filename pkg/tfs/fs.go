package tfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

type AferoFs struct {
	afero.Fs
}

func (a *AferoFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	if lstater, ok := a.Fs.(afero.Lstater); ok {
		fileInfo, _, err := lstater.LstatIfPossible(name)
		return fileInfo, true, err
	}
	return nil, false, nil
}

func (a *AferoFs) SymlinkIfPossible(target, symlink string) error {
	// Ensure the parent directory for the symlink exists
	if err := os.MkdirAll(filepath.Dir(symlink), 0755); err != nil {
		return err
	}
	return os.Symlink(target, symlink)
}

func (a *AferoFs) EvalSymlinksIfPossible(path string) (string, bool, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", false, err
	}
	return resolved, true, nil
}

// Default to real filesystem.
var AppFs = &AferoFs{afero.NewOsFs()}

// Set up a temporary test filesystem using Afero.
func initTestFS(tb testing.TB) (string, func()) {
	tb.Helper()
	tempDir := tb.TempDir()

	// Create a real OS-based filesystem using a temp directory
	AppFs = &AferoFs{Fs: afero.NewBasePathFs(afero.NewOsFs(), tempDir)}

	// Set required Viper config values
	viper.Set("terraform_file_name_prefix", "terraform_")
	viper.Set("user_bin_directory", filepath.Join(tempDir, "bin"))

	// Ensure directories exist
	if err := AppFs.MkdirAll(viper.GetString("user_bin_directory"), 0755); err != nil {
		tb.Fatalf("Failed to create bin dir: %v", err)
	}

	return tempDir, func() {
		viper.Reset()	
	}
}
