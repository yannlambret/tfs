package tfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apex/log"

	"github.com/Masterminds/semver"
)

type localCache struct {
	Directory   string
	Releases    []*release
	LastRelease *release
}

func (c *localCache) isEmpty() bool {
	return len(c.Releases) == 0
}

var cache *localCache

func init() {
	cache = new(localCache)
	// Local cache directory is "${XDG_CACHE_HOME}/tfs" by default,
	// or "${HOME}/.cache/tfs" as a fallback.
	if directory, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		cache.Directory = filepath.Join(directory, "tfs")
	} else {
		cache.Directory = filepath.Join(os.Getenv("HOME"), ".cache", "tfs")
	}

	ctx := log.WithFields(log.Fields{
		"cacheDirectory": cache.Directory,
	})

	// Check if cache directory exists, create it otherwise.
	if _, err := os.Stat(cache.Directory); os.IsNotExist(err) {
		if err := os.MkdirAll(cache.Directory, os.ModePerm); err != nil {
			ctx.WithError(err).Error("Failed to create cache directory")
			os.Exit(1)
		}
	}

	// Cache state
	files, err := filepath.Glob(filepath.Join(cache.Directory, "terraform_*"))
	if err != nil {
		ctx.WithError(err).Error("Unable to read cache directory contents")
		os.Exit(1)
	}
	if len(files) > 0 {
		versions := make([]*semver.Version, 0)

		for _, fileName := range files {
			version, err := versionFromFileName(filepath.Base(fileName))
			if err != nil {
				ctx := log.WithFields(log.Fields{
					"cacheDirectory": cache.Directory,
					"fileName":       filepath.Base(fileName),
				})
				ctx.WithError(err).Error("Invalid file name")
				os.Exit(1)
			}
			versions = append(versions, version)
		}

		sort.Sort(semver.Collection(versions))

		// Set cache releases
		cache.Releases = make([]*release, 0)

		for _, version := range versions {
			cache.Releases = append(cache.Releases, NewRelease(version).Init())
		}

		// Set cache last release
		cache.LastRelease = cache.Releases[len(cache.Releases)-1]
	}
}

// Helper to get semantic version from Terraform binary file name.
func versionFromFileName(fileName string) (*semver.Version, error) {
	return semver.NewVersion(strings.Trim(fileName, "terraform_"))
}

// Prune command can be used to wipe the whole cache.
func Prune() error {
	removed := 0
	for _, release := range cache.Releases {
		f := filepath.Join(release.CacheDirectory, release.FileName)
		ctx := log.WithFields(log.Fields{
			"version":  release.Version.String(),
			"fileName": f,
		})
		if err := os.Remove(f); err != nil {
			ctx.WithError(err).Error("Unable to remove Terraform binary")
			return err
		}
		removed++
	}
	log.Info("Removed " + fmt.Sprintf("%d", removed) + " files")
	return nil
}

// PruneUntil remove Terraform binary for the specified version and all the previous ones.
func PruneUntil(version *semver.Version) error {
	removed := 0
	// Ignoring potential errors here because we have already
	// checked that the argument is a valid semantic version.
	c, _ := semver.NewConstraint("<= " + version.String())
	for _, release := range cache.Releases {
		f := filepath.Join(release.CacheDirectory, release.FileName)
		if c.Check(release.Version) {
			ctx := log.WithFields(log.Fields{
				"version":  release.Version.String(),
				"fileName": f,
			})
			if err := os.Remove(f); err != nil {
				ctx.WithError(err).Error("Unable to remove Terraform binary")
				return err
			}
			removed++
		}
	}
	log.Info("Removed " + fmt.Sprintf("%d", removed) + " files")
	return nil
}
