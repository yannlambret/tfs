package tfs

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/viper"
)

type localCache struct {
	Directory      string
	Releases       map[string]*release
	LastRelease    *release
	CurrentRelease *release
	ActiveRelease  *release
}

// Disk location where Terraform binaries are kept.
var Cache *localCache

func init() {
	Cache = new(localCache)
}

func (c *localCache) Load() error {
	slog := slog.With(slog.String("cacheDirectory", c.Directory))

	c.Releases = make(map[string]*release)
	c.LastRelease = nil

	// Cache state.
	files, err := filepath.Glob(filepath.Join(c.Directory, viper.GetString("terraform_file_name_prefix")+"*"))
	if err != nil {
		slog.Error("Failed to load cache data", "error", err)
		return err
	}

	if len(files) > 0 {
		for _, fileName := range files {
			version, err := versionFromFileName(filepath.Base(fileName))
			if err != nil {
				slog := slog.With("fileName", filepath.Base(fileName))
				slog.Error("Invalid file name", "error", err)
				return err
			}
			// Initialize the release.
			c.Releases[version.String()] = NewRelease(version).Init()
			// Set cache most recent release.
			if c.LastRelease != nil {
				constraint, _ := semver.NewConstraint("> " + c.LastRelease.Version.String())
				if !constraint.Check(version) {
					continue
				}
			}
			c.LastRelease = c.Releases[version.String()]
		}
	}

	return nil
}

func (c *localCache) isEmpty() bool {
	return len(c.Releases) == 0
}

func (c *localCache) List() error {
	versions := make([]string, 0)
	for k := range c.Releases {
		versions = append(versions, k)
	}
	sort.Strings(versions)

	for _, v := range versions {
		r := c.Releases[v]
		if isatty.IsTerminal(os.Stderr.Fd()) {
			if r.SameAs(c.ActiveRelease) {
				color.New(color.FgHiCyan, color.Bold).Println(v + " (active)")
			} else {
				fmt.Println(v)
			}
		} else {
			slog.Info("release",
				slog.String("version", v),
				slog.Bool("isActive", r.SameAs(c.ActiveRelease)),
			)
		}
	}
	return nil
}

// Size returns the cache total size.
func (c *localCache) Size() (uint64, error) {
	var size uint64

	// Reload cache contents.
	c.Load()

	slog := slog.With(slog.String("cacheDirectory", c.Directory))

	for _, release := range c.Releases {
		releaseSize, err := release.Size()
		if err != nil {
			slog.Error("Failed to get cache size", "error", err)
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
		reclaimed uint64
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

	slog.Info(
		"Removed "+fmt.Sprintf("%d", removed)+" file(s)",
		"cacheDirectory", c.Directory,
		"cacheSize", formatSize(cacheSize),
		"reclaimedSpace", formatSize(reclaimed),
		"removed", removed,
	)

	return nil
}

// PruneUntil command removes all Terraform binary versions prior to the one specified.
func (c *localCache) PruneUntil(v *semver.Version) error {
	var (
		removed   int
		reclaimed uint64
	)

	// Ignoring potential errors here because we have already
	// checked that the argument is a valid semantic version.
	constraint, _ := semver.NewConstraint("< " + v.String())

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

	slog.Info(
		"Removed "+fmt.Sprintf("%d", removed)+" file(s)",
		"cacheDirectory", c.Directory,
		"cacheSize", formatSize(cacheSize),
		"reclaimedSpace", formatSize(reclaimed),
		"removed", removed,
	)

	return nil
}

func (c *localCache) AutoClean() {
	if !viper.GetBool("cache_auto_clean") || c.isEmpty() {
		// Feature disabled by the configuration or empty cache.
		return
	}

	// Reload cache contents.
	c.Load()

	// Maximal number of releases to keep in the cache.
	if viper.GetInt("cache_minor_version_nb") != 0 && viper.GetInt("cache_patch_version_nb") != 0 {
		// Create list of patches per minor release.
		minorReleases := make(map[string][]*semver.Version)
		for _, r := range c.Releases {
			v, _ := semver.NewVersion(fmt.Sprintf("%d.%d", r.Version.Major(), r.Version.Minor()))
			minorReleases[v.String()] = append(minorReleases[v.String()], r.Version)
		}
		for _, releases := range minorReleases {
			sort.Sort(semver.Collection(releases))
		}
		// Honoring the 'cache_minor_version_nb' configuration attribute.
		keys := make([]string, 0)
		for k := range minorReleases {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		n := len(keys) - viper.GetInt("cache_minor_version_nb")
		if n > 0 {
			toBeRemoved := keys[0:n]
			for _, version := range toBeRemoved {
				// ~1.2.3 is equivalent to >= 1.2.3, < 1.3.0
				constraint, _ := semver.NewConstraint("~" + version)
				// Remove file(s) from disk.
				for _, release := range c.Releases {
					if constraint.Check(release.Version) && !release.SameAs(c.CurrentRelease) {
						// Try to remove the file silently.
						release.Remove()
					}
				}
				// Delete releases from the map.
				delete(minorReleases, version)
			}
		}
		// Honoring the 'cache_patch_version_nb' configuration attribute.
		for _, values := range minorReleases {
			sort.Sort(semver.Collection(values))
			n := len(values) - viper.GetInt("cache_patch_version_nb")
			if n > 0 {
				toBeRemoved := values[0:n]
				for _, version := range toBeRemoved {
					if !c.Releases[version.String()].SameAs(c.CurrentRelease) {
						c.Releases[version.String()].Remove()
					}
				}
			}
		}
		return
	}

	// Default caching mode.
	cacheHistory := viper.GetInt("cache_history")

	n := len(c.Releases) - cacheHistory
	if n > 0 {
		keys := make([]string, 0)
		for k := range c.Releases {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		toBeRemoved := keys[0:n]
		for _, version := range toBeRemoved {
			if !c.Releases[version].SameAs(c.CurrentRelease) {
				c.Releases[version].Remove()
			}
		}
	}
}
