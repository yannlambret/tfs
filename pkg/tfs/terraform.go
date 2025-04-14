package tfs

import (
	"log/slog"
	"os"

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
		slog.Error("Failed to extract Terraform version from local configuration", "error", err)
		return nil, err
	}

	return v, nil
}
