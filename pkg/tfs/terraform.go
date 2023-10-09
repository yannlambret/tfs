package tfs

import (
	"log/slog"
	"os"

	"github.com/Masterminds/semver"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// getTfVersion looks for a version constraint in a set of Terraform manifest files.
func GetTfVersion() (*semver.Version, error) {
	var version string

	path, err := os.Getwd()

	if err != nil {
		slog.Error(Align(padding, "Failed to get working directory"), "error", err)
		return nil, err
	}

	slog := slog.With("path", path)

	if !tfconfig.IsModuleDir(path) {
		slog.Info(Align(padding, "TF configuration not found (are you in a module folder?)"))
		return nil, nil
	}

	module, diags := tfconfig.LoadModule(path)
	if diags.HasErrors() {
		slog.Error(Align(padding, "Failed to load TF configuration"), "error", diags.Err())
		return nil, diags.Err()
	}

	// Get Terraform semantic version for current configuration.
	if len(module.RequiredCore) != 0 {
		version = module.RequiredCore[0]
	} else {
		// No version defined in Terrafom configuration,
		// so we activate the most recent available release.
		slog.Info(Align(padding, "Version constrainst not found"))
		if !Cache.isEmpty() {
			return Cache.LastRelease.Version, nil
		} else {
			slog.Info(Align(padding, "No available TF release"))
			return nil, nil
		}
	}

	slog = slog.With(
		"path", path,
		"version", version,
	)

	v, err := semver.NewVersion(version)

	if err != nil {
		slog.Error(Align(padding, "Failed to extract TF version"), "error", err)
		return nil, err
	}

	return v, nil
}
