package tfs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"github.com/hashicorp/go-getter"
	"github.com/spf13/viper"
)

type release struct {
	Version        *semver.Version
	CacheDirectory string
	FileName       string
	URLPrefix      string
	BinaryURL      string
	ChecksumURL    string
	URL            string
}

func NewRelease(version *semver.Version) *release {
	return &release{Version: version}
}

func (r *release) Init() *release {
	// The target directory for Terraform binary.
	r.CacheDirectory = Cache.Directory

	// The local name of the Terraform binary.
	r.FileName = viper.GetString("terraform_file_name_prefix") + r.Version.String()

	// Terraform download URL prefix.
	r.URLPrefix = fmt.Sprintf("%s/terraform/%s/terraform_%s", viper.GetString("terraform_download_url"), r.Version.String(), r.Version.String())

	// Terraform binary download URL.
	r.BinaryURL = fmt.Sprintf("%s_%s_%s.zip", r.URLPrefix, runtime.GOOS, runtime.GOARCH)

	// Terraform checksum file download URL.
	r.ChecksumURL = fmt.Sprintf("%s_SHA256SUMS", r.URLPrefix)

	// Full download URL.
	r.URL = fmt.Sprintf("%s?checksum=file:%s", r.BinaryURL, r.ChecksumURL)

	return r
}

// Install downloads the required Terraform binary
// and put it in the cache directory.
func (r *release) Install() error {
	ctx := log.WithFields(log.Fields{
		"version":        r.Version.String(),
		"cacheDirectory": r.CacheDirectory,
		"fileName":       r.FileName,
		"binaryURL":      r.BinaryURL,
	})

	// Check if the desired Terraform binary is already
	// installed, download it otherwise.
	path := filepath.Join(r.CacheDirectory, r.FileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		ctx.Info(Align(padding, "Downloading Terraform"))
		if err := getter.GetFile(path, r.URL); err != nil {
			ctx.WithError(err).Error(Align(padding, "Download failed"))
			return err
		}
	}

	return nil
}

// Activate creates the symbolic link in the user path that
// points to the desired Terraform binary.
func (r *release) Activate() error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Error(Align(padding, "Failed to get user home directory"))
		return err
	}

	userBinDir := filepath.Join(userHomeDir, ".local", "bin")
	target := filepath.Join(r.CacheDirectory, r.FileName)
	symlink := filepath.Join(userBinDir, "terraform")

	ctx := log.WithFields(log.Fields{
		"userBinDir": userBinDir,
	})

	if _, err := os.Stat(userBinDir); os.IsNotExist(err) {
		ctx.Info(Align(padding, "Creating bin directory"))
		if err := os.MkdirAll(userBinDir, os.ModePerm); err != nil {
			ctx.WithError(err).Error(Align(padding, "Operation failed"))
			return err
		}
	}

	ctx = log.WithFields(log.Fields{
		"version": r.Version.String(),
		"target":  target,
		"symlink": symlink,
	})

	// Check if the desired version is already active.
	if path, _ := filepath.EvalSymlinks(symlink); path == target {
		ctx.Info(Align(padding, "Version is already active"))
		return nil
	}

	// Remove the link if it exists.
	if _, err := os.Lstat(symlink); err == nil {
		os.Remove(symlink)
	}

	// Create the symbolic link.
	if err := os.Symlink(target, symlink); err != nil {
		ctx.WithError(err).Error(Align(padding, "Failed to create symlink"))
		return err
	}
	ctx.Info(Align(padding, "New active version"))

	return nil
}

// Remove deletes a specific Terraform binary from the local cache.
func (r *release) Remove() error {
	f := filepath.Join(r.CacheDirectory, r.FileName)
	ctx := log.WithFields(log.Fields{
		"version":  r.Version.String(),
		"fileName": f,
	})
	if err := os.Remove(f); err != nil {
		ctx.WithError(err).Error(Align(padding, "Failed to remove TF binary"))
		return err
	}

	return nil
}

// Size function returns the size of the Terraform binary.
func (r *release) Size() (float64, error) {
	f := filepath.Join(Cache.Directory, r.FileName)
	fi, err := os.Stat(f)
	ctx := log.WithFields(log.Fields{
		"version":  r.Version.String(),
		"fileName": f,
	})
	if err != nil {
		ctx.WithError(err).Error(Align(padding, "Failed to get TF binary information"))
		return 0, err
	}

	return float64(fi.Size()), nil
}
