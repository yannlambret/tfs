package tfs

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/spf13/viper"
)

type release struct {
	parentCache *LocalCache
	Version     *version.Version
	fileName    string
}

// Install downloads the required Terraform binary
// and put it in the cache directory.
func (r *release) Install() error {
	logger := slog.With(
		"cacheDirectory", r.parentCache.directory,
		"version", r.Version.String(),
		"fileName", r.fileName,
	)

	// Check if the desired Terraform binary is already
	// installed, download it otherwise.
	targetPath := filepath.Join(r.parentCache.directory, r.fileName)

	// Ensure parent cache directory exists.
	if err := AppFs.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		logger.Error("Failed to create cache directory", "error", err)
		return err
	}

	if _, err := AppFs.Stat(targetPath); os.IsNotExist(err) {
		ctx := context.Background()
		i := install.NewInstaller()
		defer i.Remove(ctx)

		logger.Info("Downloading Terraform")

		srcPath, err := i.Install(ctx, []src.Installable{
			&releases.ExactVersion{
				Product: product.Terraform,
				Version: r.Version,
			},
		})
		if err != nil {
			logger.Error("Download failed", "error", err)
			return err
		}
		// Move downloaded file.
		b, err := os.ReadFile(srcPath)
		if err != nil {
			logger.Error("Unable to read downloaded file", "error", err, "srcPath", srcPath)
			return err
		}
		err = os.WriteFile(targetPath, b, os.ModePerm)
		if err != nil {
			logger.Error("Unable to move downloaded file to cache", "error", err, "targetPath", targetPath)
			return err
		}
	}

	// Keep track of the current release for we don't
	// want the last downloaded version to be removed
	// by the cache cleanup routine.
	r.parentCache.currentRelease = r

	return nil
}

// Activate creates the symbolic link in the user path that
// points to the desired Terraform binary.
func (r *release) Activate() error {
	var (
		userBinDir = viper.GetString("user_bin_directory")
		symlink    = filepath.Join(userBinDir, "terraform")
		target     = filepath.Join(r.parentCache.directory, r.fileName)
	)

	userBinLogger := slog.With(
		"userBinDir", userBinDir,
	)

	if _, err := AppFs.Stat(userBinDir); os.IsNotExist(err) {
		userBinLogger.Info("Creating user local bin directory")
		if err := AppFs.MkdirAll(userBinDir, os.ModePerm); err != nil {
			userBinLogger.Error("Operation failed", "error", err)
			return err
		}
		userBinLogger.Warn("Make sure to add local bin directory to PATH environment variable")
	}

	// Create a new logger for activation logs (doesn't include userBinDir).
	activateLogger := slog.With(
		"version", r.Version.String(),
		"target", target,
		"symlink", symlink,
	)

	// Check if the desired version is already active.
	if r.SameAs(r.parentCache.activeRelease) {
		activateLogger.Info("Version is already active")
		return nil
	}

	// Remove the link if it exists.
	if _, b, err := AppFs.LstatIfPossible(symlink); !b {
		activateLogger.Warn("The operating system does not seem to support `os.Lstat`", "error", err)
	} else if err == nil {
		AppFs.Remove(symlink)
	}

	// Create the symbolic link.
	if err := AppFs.SymlinkIfPossible(target, symlink); err != nil {
		activateLogger.Error("Failed to create symlink", "error", "err")
		return err
	}

	r.parentCache.activeRelease = r
	activateLogger.Info("New active version")

	return nil
}

// Remove deletes a specific Terraform binary from the local cache.
func (r *release) Remove() error {
	var (
		userBinDir = viper.GetString("user_bin_directory")
		symlink    = filepath.Join(userBinDir, "terraform")
		target     = filepath.Join(r.parentCache.directory, r.fileName)
	)

	logger := slog.With(
		"version", r.Version.String(),
		"fileName", target,
	)

	// Check if we should also remove the symbolic link.
	if path, ok, _ := AppFs.EvalSymlinksIfPossible(symlink); ok && path == target {
		AppFs.Remove(symlink)
	}

	if err := AppFs.Remove(target); err != nil {
		logger.Error("Failed to remove Terraform binary", "error", err)
		return err
	}

	return nil
}

// Size function returns the size of the Terraform binary.
func (r *release) Size() (uint64, error) {
	target := filepath.Join(r.parentCache.directory, r.fileName)

	logger := slog.With(
		"version", r.Version.String(),
		"fileName", target,
	)

	fi, err := AppFs.Stat(target)

	if err != nil {
		logger.Error("Failed to get Terraform binary information", "error", err)
		return 0, err
	}

	return uint64(fi.Size()), nil
}

// SameAs compares the current release and the given release.
func (r *release) SameAs(ref *release) bool {
	return reflect.DeepEqual(r, ref)
}
