package tfs

import (
	"os"

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
	return os.Symlink(target, symlink)
}

// We are using this as a wrapper to the regular 'os' package,
// so that we can switch the implementation to an in-memory
// filesystem when testing.
var AppFs = &AferoFs{afero.NewOsFs()}
