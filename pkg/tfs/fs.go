package tfs

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
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
	// Ensure the parent directory for the symlink exists.
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
