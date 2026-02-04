package tfs

import (
	"log/slog"
	"os"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// GetTfVersion looks for a version constraint in Terraform manifest files
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
