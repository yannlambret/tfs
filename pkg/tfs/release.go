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
	Version        *version.Version
	CacheDirectory string
	FileName       string
}

func NewRelease(v *version.Version) *release {
	return &release{Version: v}
}

func (r *release) Init() *release {
	var (
		userBinDir = viper.GetString("user_bin_directory")
		symlink    = filepath.Join(userBinDir, "terraform")
		target, _  = filepath.EvalSymlinks(symlink)
	)

	// The target directory for Terraform binary.
	r.CacheDirectory = Cache.Directory

	// The local name of the Terraform binary.
	r.FileName = viper.GetString("terraform_file_name_prefix") + r.Version.String()

	// Already installed and active?
	if target == filepath.Join(r.CacheDirectory, r.FileName) {
		Cache.ActiveRelease = r
	}

	return r
}

// Install downloads the required Terraform binary
// and put it in the cache directory.
func (r *release) Install() error {
	slog := slog.With(
		"version", r.Version.String(),
		"cacheDirectory", r.CacheDirectory,
		"fileName", r.FileName,
	)

	// Check if the desired Terraform binary is already
	// installed, download it otherwise.
	targetPath := filepath.Join(r.CacheDirectory, r.FileName)

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		ctx := context.Background()
		i := install.NewInstaller()

		slog.Info("Downloading Terraform")

		srcPath, err := i.Install(ctx, []src.Installable{
			&releases.ExactVersion{
				Product: product.Terraform,
				Version: r.Version,
			},
		})
		if err != nil {
			slog.Error("Download failed", "error", err)
			return err
		}
		// Move downloaded file.
		b, err := os.ReadFile(srcPath)
		if err != nil {
			slog.Error("Unable to read downloaded file", "error", err, "srcPath", srcPath)
			return err
		}
		err = os.WriteFile(targetPath, b, os.ModePerm)
		if err != nil {
			slog.Error("Unable to move downloaded file to cache", "error", err, "targetPath", targetPath)
			return err
		}
	}

	// Keep track of the current release for we don't
	// want the last downloaded version to be removed
	// by the cache cleanup routine.
	Cache.CurrentRelease = r

	return nil
}

// Activate creates the symbolic link in the user path that
// points to the desired Terraform binary.
func (r *release) Activate() error {
	var (
		userBinDir = viper.GetString("user_bin_directory")
		symlink    = filepath.Join(userBinDir, "terraform")
		target     = filepath.Join(r.CacheDirectory, r.FileName)
	)

	slog := slog.With(
		"userBinDir", userBinDir,
	)

	if _, err := os.Stat(userBinDir); os.IsNotExist(err) {
		slog.Info("Creating user local bin directory")
		if err := os.MkdirAll(userBinDir, os.ModePerm); err != nil {
			slog.Error("Operation failed", "error", err)
			return err
		}
		slog.Warn("Make sure to add local bin directory to PATH environment variable")
	}

	slog = slog.With(
		"version", r.Version.String(),
		"target", target,
		"symlink", symlink,
	)

	// Check if the desired version is already active.
	if r.SameAs(Cache.ActiveRelease) {
		slog.Info("Version is already active")
		return nil
	}

	// Remove the link if it exists.
	if _, err := os.Lstat(symlink); err == nil {
		os.Remove(symlink)
	}

	// Create the symbolic link.
	if err := os.Symlink(target, symlink); err != nil {
		slog.Error("Failed to create symlink", "error", "err")
		return err
	}

	Cache.ActiveRelease = r
	slog.Info("New active version")

	return nil
}

// Remove deletes a specific Terraform binary from the local cache.
func (r *release) Remove() error {
	var (
		userBinDir = viper.GetString("user_bin_directory")
		symlink    = filepath.Join(userBinDir, "terraform")
		target     = filepath.Join(r.CacheDirectory, r.FileName)
	)

	slog := slog.With(
		"version", r.Version.String(),
		"fileName", target,
	)

	// Check if we should also remove the symbolic link.
	if path, _ := filepath.EvalSymlinks(symlink); path == target {
		os.Remove(symlink)
	}

	if err := os.Remove(target); err != nil {
		slog.Error("Failed to remove Terraform binary", "error", err)
		return err
	}

	return nil
}

// Size function returns the size of the Terraform binary.
func (r *release) Size() (uint64, error) {
	target := filepath.Join(Cache.Directory, r.FileName)

	slog := slog.With(
		"version", r.Version.String(),
		"fileName", target,
	)

	fi, err := os.Stat(target)

	if err != nil {
		slog.Error("Failed to get Terraform binary information", "error", err)
		return 0, err
	}

	return uint64(fi.Size()), nil
}

// SameAs compares the current release and the given release.
func (r *release) SameAs(ref *release) bool {
	return reflect.DeepEqual(r, ref)
}
