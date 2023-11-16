package tfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func TestNewRelease(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	v, _ := version.NewVersion("1.10.0")
	cache := NewLocalCache(cacheDir)
	release := cache.NewRelease(v)

	expectedFileName := getFilePrefix() + v.String()

	if release.Version.String() != v.String() {
		t.Errorf("Expected version %s, got %s", v.String(), release.Version.String())
	}
	if release.fileName != expectedFileName {
		t.Errorf("Expected fileName %s, got %s", expectedFileName, release.fileName)
	}
	if release.parentCache != cache {
		t.Errorf("Expected parentCache to match cache instance")
	}
}

func TestReleaseActivateCreatesSymlink(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	binDir := viper.GetString("user_bin_directory")
	tfVersion := "1.10.0"

	binaryPath := filepath.Join(cacheDir, getFilePrefix()+tfVersion)
	writeTestFile(t, binaryPath, []byte("dummy content"))

	v, _ := version.NewVersion(tfVersion)
	cache := NewLocalCache(cacheDir)
	release := cache.NewRelease(v)

	if err := release.Activate(); err != nil {
		t.Fatalf("Activate() failed: %v", err)
	}

	symlinkPath := filepath.Join(binDir, "terraform")
	fi, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Symlink not found: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("Expected a symlink, got regular file")
	}

	resolved, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to resolve symlink: %v", err)
	}
	if resolved != binaryPath {
		t.Errorf("Symlink points to %q, expected %q", resolved, binaryPath)
	}
}

func TestReleaseRemoveDeletesBinaryAndSymlink(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	v, _ := version.NewVersion("1.10.0")
	cache := NewLocalCache(cacheDir)
	release := cache.NewRelease(v)

	releasePath := filepath.Join(cacheDir, release.fileName)
	writeTestFile(t, releasePath, []byte("dummy content"))

	symlink := filepath.Join(viper.GetString("user_bin_directory"), "terraform")
	if err := AppFs.SymlinkIfPossible(releasePath, symlink); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	if err := release.Remove(); err != nil {
		t.Fatalf("release.Remove() failed: %v", err)
	}

	if exists, _ := afero.Exists(AppFs, releasePath); exists {
		t.Errorf("Expected release binary %s to be deleted", releasePath)
	}
	
	if exists, _ := afero.Exists(AppFs, symlink); exists {
		t.Errorf("Expected symlink %s to be deleted", symlink)
	}
}
