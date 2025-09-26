package tfs

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"regexp"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// getTfVersion looks for a version constraint in a set of Terraform manifest files.
func GetTfVersion() (*version.Version, error) {
	var tfVersion string

	path, err := os.Getwd()

	if err != nil {
		slog.Error("Failed to get working directory", "error", err)
		return nil, err
	}

	slog := slog.With("path", path)

	if !tfconfig.IsModuleDir(path) {
		slog.Info("Terraform configuration not found (are you in a module folder?)")
		return nil, nil
	}

	module, diags := tfconfig.LoadModule(path)
	if diags.HasErrors() {
		slog.Error("Failed to load Terraform configuration", "error", diags.Err())
		return nil, diags.Err()
	}

	// Get Terraform semantic version for current configuration.
	if len(module.RequiredCore) != 0 {
		tfVersion = module.RequiredCore[0]
	} else {
		// No version defined in Terrafom configuration.
		return nil, nil
	}

	slog = slog.With(
		"path", path,
		"version", tfVersion,
	)

	v, err := version.NewVersion(tfVersion)
	if err != nil {
		v, err = parseVersionConstraint(tfVersion)
		if err != nil {
		slog.Error("Failed to extract Terraform version from local configuration", "error", err)
		return nil, err
		}
	}

	return v, nil
}
func parseVersionConstraint(constraint string) (*version.Version, error) {
    // Regex to extract version numbers from various constraint formats
    re := regexp.MustCompile(`[~>=<!]*\s*([0-9]+(?:\.[0-9]+)*(?:\.[0-9]+)*)`)
    
    matches := re.FindStringSubmatch(strings.TrimSpace(constraint))
    if len(matches) < 2 {
        return nil, fmt.Errorf("no version found in constraint: %s", constraint)
    }
    
    versionStr := matches[1]
    
    // Normalize version (ensure it has major.minor.patch)
    parts := strings.Split(versionStr, ".")
    for len(parts) < 3 {
        parts = append(parts, "0")
    }
    normalizedVersion := strings.Join(parts[:3], ".")
    
    return version.NewVersion(normalizedVersion)
}