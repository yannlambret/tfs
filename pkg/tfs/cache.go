package tfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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

func (c *localCache) Load() error {
	ctx := log.WithFields(log.Fields{
		"cacheDirectory": c.Directory,
	})

	versions := make([]*semver.Version, 0)
	c.Releases = make([]*release, 0)

	// Cache state.
	files, err := filepath.Glob(filepath.Join(c.Directory, tfFileNamePrefix+"*"))
	if err != nil {
		ctx.WithError(err).Error("Failed to load cache data")
		return err
	}

	if len(files) > 0 {
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

// Size returns the total cache size.
func (c *localCache) Size() (float64, error) {
	var size float64

	// Reload cache contents.
	c.Load()

	ctx := log.WithFields(log.Fields{
		"cacheDirectory": c.Directory,
	})

	for _, release := range c.Releases {
		releaseSize, err := release.Size()
		if err != nil {
			ctx.WithError(err).Error("Failed to get cache size")
			return 0, err
		}
		size += releaseSize
	}

	return size, nil
}

func (c *localCache) List() error {
	return nil
}

// Prune command can be used to wipe the whole cache.
func (c *localCache) Prune() error {
	var (
		removed   int
		reclaimed float64
	)

	for _, release := range c.Releases {
		releaseSize, err := release.Size()
		if err != nil {
			return err
		}
		if err := release.Remove(); err != nil {
			return err
		}
		removed++
		reclaimed += releaseSize
	}

	cacheSize, err := c.Size()
	if err != nil {
		return err
	}

	ctx := log.WithFields(log.Fields{
		"cacheDirectory": c.Directory,
		"cacheSize":      formatSize(cacheSize),
		"reclaimedSpace": formatSize(reclaimed),
	})

	ctx.Info("Removed " + fmt.Sprintf("%d", removed) + " file(s)")

	return nil
}

// PruneUntil command removes all Terraform binary versions prior to the one specified.
func (c *localCache) PruneUntil(version *semver.Version) error {
	var (
		removed   int
		reclaimed float64
	)

	// Ignoring potential errors here because we have already
	// checked that the argument is a valid semantic version.
	constraint, _ := semver.NewConstraint("< " + version.String())

	for _, release := range c.Releases {
		if constraint.Check(release.Version) {
			releaseSize, err := release.Size()
			if err != nil {
				return err
			}
			if err := release.Remove(); err != nil {
				return err
			}
			removed++
			reclaimed += releaseSize
		}
	}

	cacheSize, err := c.Size()
	if err != nil {
		return err
	}

	ctx := log.WithFields(log.Fields{
		"cacheDirectory": c.Directory,
		"cacheSize":      formatSize(cacheSize),
		"reclaimedSpace": formatSize(reclaimed),
	})

	ctx.Info("Removed " + fmt.Sprintf("%d", removed) + " file(s)")

	return nil
}
