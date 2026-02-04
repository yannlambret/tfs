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
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// LocalCache holds information about downloaded Terraform releases.
type LocalCache struct {
	directory      string
	releases       map[string]*release
	activeRelease  *release
	currentRelease *release
	LastRelease    *release // public
}

// NewLocalCache creates the LocalCache with the given directory.
func NewLocalCache(directory string) *LocalCache {
	return &LocalCache{
		directory: directory,
		releases:  make(map[string]*release),
	}
}

// NewRelease creates a new cached release.
func (c *LocalCache) NewRelease(v *version.Version) *release {
	r := &release{
		Version:     v, // public
		fileName:    viper.GetString("terraform_file_name_prefix") + v.String(),
		parentCache: c,
	}

	// Check if this release is the active one.
	userBinDir := viper.GetString("user_bin_directory")
	symlink := filepath.Join(userBinDir, "terraform")

	if target, ok, _ := AppFs.EvalSymlinksIfPossible(symlink); ok && target == filepath.Join(c.directory, r.fileName) {
		c.activeRelease = r
	}

	return r
}

// Load builds the cache state based on the Terraform versions
// that have already been downloaded in the cache directory.
func (c *LocalCache) Load() error {
	logger := slog.With(slog.String("cacheDirectory", c.directory))

	c.releases = make(map[string]*release)
	c.LastRelease = nil

	// Cache state.
	files, err := afero.Glob(AppFs, filepath.Join(c.directory, viper.GetString("terraform_file_name_prefix")+"*"))
	if err != nil {
		logger.Error("Failed to load cache data", "error", err)
		return err
	}

	for _, fileName := range files {
		v, err := versionFromFileName(filepath.Base(fileName))
		if err != nil {
			fileLogger := logger.With("fileName", filepath.Base(fileName))
			fileLogger.Error("Invalid file name", "error", err)
			return err
		}
		r := c.NewRelease(v)
		c.releases[v.String()] = r

		// Update last release based on version order.
		if c.LastRelease != nil {
			constraint, _ := newConstraintExtended(">" + c.LastRelease.Version.String())
			if !constraint.Check(v) {
				continue
			}
		}
		c.LastRelease = r
	}

	return nil
}

// IsEmpty allows to check if the cache is empty.
func (c *LocalCache) IsEmpty() bool {
	return len(c.releases) == 0
}

// List command displays the contents of the local cache.
func (c *LocalCache) List() error {
	versions := make([]*version.Version, 0, len(c.releases))
	for _, r := range c.releases {
		versions = append(versions, r.Version)
	}
	sort.Sort(version.Collection(versions))

	for _, v := range versions {
		r := c.releases[v.String()]
		if isatty.IsTerminal(os.Stderr.Fd()) {
			if r.SameAs(c.activeRelease) {
				color.New(color.FgHiCyan, color.Bold).Println(v.String() + " (active)")
			} else {
				fmt.Println(v.String())
			}
		} else {
			slog.Info("release",
				slog.String("version", v.String()),
				slog.Bool("isActive", r.SameAs(c.activeRelease)),
			)
		}
	}

	return nil
}

// Size returns the cache total size.
func (c *LocalCache) Size() (uint64, error) {
	var size uint64

	// Reload cache contents.
	c.Load()

	logger := slog.With(slog.String("cacheDirectory", c.directory))

	for _, release := range c.releases {
		releaseSize, err := release.Size()
		if err != nil {
			logger.Error("Failed to get cache size", "error", err)
			return 0, err
		}
		size += releaseSize
	}

	return size, nil
}

// Prune command can be used to wipe the whole cache.
func (c *LocalCache) Prune() error {
	var (
		removed   int
		reclaimed uint64
	)

	for _, release := range c.releases {
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
		"cacheDirectory", c.directory,
		"cacheSize", formatSize(cacheSize),
		"reclaimedSpace", formatSize(reclaimed),
		"removed", removed,
	)

	return nil
}

// PruneUntil command removes all Terraform binary versions prior to the one specified.
func (c *LocalCache) PruneUntil(v *version.Version) error {
	var (
		removed   int
		reclaimed uint64
	)

	// Ignoring potential errors here because we have already
	// checked that the argument is a valid semantic version.
	constraint, _ := newConstraintExtended("<" + v.String())

	for _, release := range c.releases {
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
		"cacheDirectory", c.directory,
		"cacheSize", formatSize(cacheSize),
		"reclaimedSpace", formatSize(reclaimed),
		"removed", removed,
	)

	return nil
}

// GetTfVersion looks for a version constraint in Terraform manifest files
// and resolves it to a specific version from the cache.
func (c *LocalCache) GetTfVersion() (*version.Version, error) {
	constraintStr, err := GetTfVersionConstraint()
	if err != nil {
		return nil, err
	}

	if constraintStr == "" {
		// No constraint found.
		return nil, nil
	}

	logger := slog.With(
		"constraint", constraintStr,
	)

	// First, try to parse as a plain version (e.g., "1.14.1" or "= 1.14.1").
	// This handles the common case of exact version specifications.
	if v, err := version.NewVersion(constraintStr); err == nil {
		// It's a plain version string, return it directly.
		logger.Info("Found version requirement", "version", v.String())
		return v, nil
	}

	// Parse as a constraint (supports ~> operator and ranges).
	constraint, err := newConstraintExtended(constraintStr)
	if err != nil {
		logger.Error("Failed to parse Terraform version constraint", "error", err)
		return nil, err
	}

	// Find the best matching version from the cache.
	var bestMatch *version.Version
	for _, release := range c.releases {
		if constraint.Check(release.Version) {
			if bestMatch == nil || release.Version.GreaterThan(bestMatch) {
				bestMatch = release.Version
			}
		}
	}

	if bestMatch != nil {
		logger.Info("Resolved version constraint", "version", bestMatch.String())
		return bestMatch, nil
	}

	// No matching version in cache. Try to extract a suitable version to download.
	// For pessimistic constraints like "~> 1.12.0", extract the base version.
	// For range constraints like ">= 1.12.0", extract the minimum version.
	candidateVer := extractVersionFromConstraint(constraintStr)
	if candidateVer != nil {
		logger.Info("No cached match, will download base version from constraint", "version", candidateVer.String())
		return candidateVer, nil
	}

	// We can't determine an exact version to download.
	logger.Error("No cached version matches constraint and cannot determine exact version to download")
	return nil, fmt.Errorf("no cached version satisfies constraint %s; please specify an exact version or install a matching version first", constraintStr)
}

func (c *LocalCache) AutoClean() {
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

		for _, r := range c.releases {
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
				constraint, _ := newConstraintExtended(fmt.Sprintf("~> %s", v.String()))
				for _, release := range c.releases {
					if constraint.Check(release.Version) && !release.SameAs(c.currentRelease) {
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
					if r, ok := c.releases[v.String()]; ok && !r.SameAs(c.currentRelease) {
						r.Remove()
					}
				}
			}
		}

		return
	}

	// Default caching mode.
	cacheHistory := viper.GetInt("cache_history")

	if n := len(c.releases) - cacheHistory; n > 0 {
		versions := make([]*version.Version, 0, len(c.releases))
		for _, r := range c.releases {
			versions = append(versions, r.Version)
		}
		sort.Sort(version.Collection(versions))
		for _, v := range versions[:n] {
			if r, ok := c.releases[v.String()]; ok && !r.SameAs(c.currentRelease) {
				r.Remove()
			}
		}
	}
}
