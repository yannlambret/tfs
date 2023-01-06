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

// Disk location where Terraform binaries are kept.
var Cache *localCache

func init() {
	Cache = new(localCache)
	// Local cache directory is "${XDG_CACHE_HOME}/tfs" by default,
	// or "${HOME}/.cache/tfs" as a fallback.
	if directory, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		Cache.Directory = filepath.Join(directory, "tfs")
	} else {
		Cache.Directory = filepath.Join(os.Getenv("HOME"), ".cache", "tfs")
	}
}

// Helper to get semantic version from Terraform binary file name.
func versionFromFileName(fileName string) (*semver.Version, error) {
	return semver.NewVersion(strings.ReplaceAll(fileName, tfFileNamePrefix, ""))
}

func (c *localCache) Init() error {
	ctx := log.WithFields(log.Fields{
		"cacheDirectory": c.Directory,
	})

	// Cache state.
	files, err := filepath.Glob(filepath.Join(c.Directory, tfFileNamePrefix+"*"))
	if err != nil {
		ctx.WithError(err).Error("Unable to read cache directory contents")
		return err
	}
	if len(files) > 0 {
		versions := make([]*semver.Version, 0)

		for _, fileName := range files {
			version, err := versionFromFileName(filepath.Base(fileName))
			if err != nil {
				ctx := log.WithFields(log.Fields{
					"cacheDirectory": c.Directory,
					"fileName":       filepath.Base(fileName),
				})
				ctx.WithError(err).Error("Invalid file name")
				return err
			}
			versions = append(versions, version)
		}

		sort.Sort(semver.Collection(versions))

		// Set cache releases.
		c.Releases = make([]*release, 0)

		for _, version := range versions {
			c.Releases = append(c.Releases, NewRelease(version).Init())
		}

		c.LastRelease = c.Releases[len(c.Releases)-1]
	}
	return nil
}

func (c *localCache) isEmpty() bool {
	return len(c.Releases) == 0
}

// Prune command can be used to wipe the whole cache.
func (c *localCache) Prune() error {
	removed := 0
	for _, release := range c.Releases {
		if err := release.Remove(); err != nil {
			return err
		}
		removed++
	}
	log.Info("Removed " + fmt.Sprintf("%d", removed) + " files")
	return nil
}

// PruneUntil remove Terraform binary for the specified version and all the previous ones.
func (c *localCache) PruneUntil(version *semver.Version) error {
	removed := 0
	// Ignoring potential errors here because we have already
	// checked that the argument is a valid semantic version.
	constraint, _ := semver.NewConstraint("<= " + version.String())
	for _, release := range c.Releases {
		if constraint.Check(release.Version) {
			if err := release.Remove(); err != nil {
				return err
			}
			removed++
		}
	}
	log.Info("Removed " + fmt.Sprintf("%d", removed) + " files")
	return nil
}
