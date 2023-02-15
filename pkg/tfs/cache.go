package tfs

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"github.com/spf13/viper"
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
}

func (c *localCache) Load() error {
	ctx := log.WithFields(log.Fields{
		"cacheDirectory": c.Directory,
	})

	versions := make([]*semver.Version, 0)
	c.Releases = make([]*release, 0)

	// Cache state.
	files, err := filepath.Glob(filepath.Join(c.Directory, viper.GetString("terraform_file_name_prefix")+"*"))
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

func (c *localCache) AutoClean() {
	if !viper.GetBool("cache_auto_clean") {
		// Feature disabled by the configuration.
		return
	}

	// Reload cache contents.
	c.Load()

	// Maximal number of releases to keep in the cache.
	cacheHistory := viper.GetInt("cache_history")

	n := len(c.Releases) - cacheHistory
	if n > 0 {
		toBeRemoved := c.Releases[0:n]
		for _, release := range toBeRemoved {
			// Try to remove the file silently.
			release.Remove()
		}
	}
}

func (c *localCache) isEmpty() bool {
	return len(c.Releases) == 0
}

// Size returns the cache total size.
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
