package tfs

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// GetTfVersionConstraint looks for a version constraint in Terraform manifest files
// and returns the constraint string.
func GetTfVersionConstraint() (string, error) {
	path, err := os.Getwd()

	if err != nil {
		slog.Error("Failed to get working directory", "error", err)
		return "", err
	}

	logger := slog.With("path", path)

	if !tfconfig.IsModuleDir(path) {
		logger.Info("Terraform configuration not found (are you in a module folder?)")
		return "", nil
	}

	module, diags := tfconfig.LoadModule(path)
	if diags.HasErrors() {
		logger.Error("Failed to load Terraform configuration", "error", diags.Err())
		return "", diags.Err()
	}

	// Get Terraform version constraint from current configuration.
	if len(module.RequiredCore) != 0 {
		return module.RequiredCore[0], nil
	}

	// No version defined in Terraform configuration.
	return "", nil
}

// ResolveVersion resolves a constraint string to a specific version, checking
// against the provided cached versions. Returns nil, nil if constraintStr is empty.
func ResolveVersion(constraintStr string, cachedVersions []*version.Version) (*version.Version, error) {
	if constraintStr == "" {
		return nil, nil
	}

	logger := slog.With("constraint", constraintStr)

	// Try to parse as a plain version (e.g., "1.14.1" or "= 1.14.1").
	if v, err := version.NewVersion(constraintStr); err == nil {
		logger.Info("Found version requirement", "version", v.String())
		return v, nil
	}

	// Parse as a constraint (go-version natively supports ~>, >=, <, etc.).
	constraint, err := version.NewConstraint(constraintStr)
	if err != nil {
		logger.Error("Failed to parse Terraform version constraint", "error", err)
		return nil, err
	}

	// Find the best matching version from the cache.
	var bestMatch *version.Version
	for _, v := range cachedVersions {
		if constraint.Check(v) {
			if bestMatch == nil || v.GreaterThan(bestMatch) {
				bestMatch = v
			}
		}
	}

	if bestMatch != nil {
		logger.Info("Resolved version constraint", "version", bestMatch.String())
		return bestMatch, nil
	}

	return nil, fmt.Errorf("no cached version satisfies constraint %q; run 'tfs <version>' to install the version you need", constraintStr)
}
