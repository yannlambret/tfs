package tfs

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
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

// Load builds the cache state based on the Terraform versions
// that have already been downloaded in the cache directory.
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
			v, err := versionFromFileName(filepath.Base(fileName))
			if err != nil {
				slog := slog.With("fileName", filepath.Base(fileName))
				slog.Error("Invalid file name", "error", err)
				return err
			}
			// Initialize the release.
			c.Releases[v.String()] = NewRelease(v).Init()
			// Set cache most recent release.
			if c.LastRelease != nil {
				constraint, _ := version.NewConstraint(">" + c.LastRelease.Version.String())
				if !constraint.Check(v) {
					continue
				}
			}
			c.LastRelease = c.Releases[v.String()]
		}
	}

	return nil
}

// isEmpty allows to check if the cache is empty.
func (c *localCache) isEmpty() bool {
	return len(c.Releases) == 0
}

// List command displays the contents of the local cache.
func (c *localCache) List() error {
	versions := make([]*version.Version, 0, len(c.Releases))
	for _, r := range c.Releases {
		versions = append(versions, r.Version)
	}
	sort.Sort(version.Collection(versions))

	for _, v := range versions {
		r := c.Releases[v.String()]
		if isatty.IsTerminal(os.Stderr.Fd()) {
			if r.SameAs(c.ActiveRelease) {
				color.New(color.FgHiCyan, color.Bold).Println(v.String() + " (active)")
			} else {
				fmt.Println(v.String())
			}
		} else {
			slog.Info("release",
				slog.String("version", v.String()),
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
func (c *localCache) PruneUntil(v *version.Version) error {
	var (
		removed   int
		reclaimed uint64
	)

	// Ignoring potential errors here because we have already
	// checked that the argument is a valid semantic version.
	constraint, _ := version.NewConstraint("<" + v.String())

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
	// Reload cache contents.
	c.Load()

	if !viper.GetBool("cache_auto_clean") {
		// Feature disabled.
		return
	}

	minorLimit := viper.GetInt("cache_minor_version_nb")
	patchLimit := viper.GetInt("cache_patch_version_nb")

	/// Keep N minor versions and M patches per minor version.
	if minorLimit > 0 && patchLimit > 0 {
		// Making groups for each minor version.
		minorReleases := make(map[string][]*version.Version)
		minorKeysSet := make(map[string]struct{})

		for _, r := range c.Releases {
			segments := r.Version.Segments()
			minorKey := fmt.Sprintf("%d.%d", segments[0], segments[1])
			minorReleases[minorKey] = append(minorReleases[minorKey], r.Version)
			minorKeysSet[minorKey] = struct{}{}
		}

		// Sort patch versions in each group.
		for _, releases := range minorReleases {
			sort.Sort(version.Collection(releases))
		}

		// Sort minor versions.
		minorKeys := make([]*version.Version, 0, len(minorKeysSet))
		for k := range minorKeysSet {
			v, _ := version.NewVersion(k)
			minorKeys = append(minorKeys, v)
		}
		sort.Sort(version.Collection(minorKeys))

		// Drop the oldest minor releases if needed.
		if n := len(minorKeys) - viper.GetInt("cache_minor_version_nb"); n > 0 {
			for _, v := range minorKeys[:n] {
				constraint, _ := version.NewConstraint(fmt.Sprintf("~>%s", v.String()))
				for _, release := range c.Releases {
					if constraint.Check(release.Version) && !release.SameAs(c.CurrentRelease) {
						release.Remove()
					}
				}
				delete(minorReleases, v.String())
			}
		}
		// Drop the oldest patch releases if needed.
		for _, versions := range minorReleases {
			if n := len(versions) - viper.GetInt("cache_patch_version_nb"); n > 0 {
				for _, v := range versions[:n] {
					if r, ok := c.Releases[v.String()]; ok && !r.SameAs(c.CurrentRelease) {
						r.Remove()
					}
				}
			}
		}

		return
	}

	// Default caching mode.
	cacheHistory := viper.GetInt("cache_history")

	if n := len(c.Releases) - cacheHistory; n > 0 {
		versions := make([]*version.Version, 0, len(c.Releases))
		for _, r := range c.Releases {
			versions = append(versions, r.Version)
		}
		sort.Sort(version.Collection(versions))
		for _, v := range versions[:n] {
			if r, ok := c.Releases[v.String()]; ok && !r.SameAs(c.CurrentRelease) {
				r.Remove()
			}
		}
	}
}
