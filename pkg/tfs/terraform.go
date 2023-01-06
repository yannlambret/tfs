package tfs

import (
	"os"

	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

const (
	tfDownloadURL    = "https://releases.hashicorp.com"
	tfFileNamePrefix = "terraform_"
)

// getTfVersion looks for a version constraint in a set of Terraform manifest files.
func GetTfVersion() (*semver.Version, error) {
	var version string

	path, err := os.Getwd()

	if err != nil {
		log.WithError(err).Error("Unable to get working directory")
		return nil, err
	}

	ctx := log.WithFields(log.Fields{
		"path": path,
	})

	if !tfconfig.IsModuleDir(path) {
		ctx.Info("Terraform configuration not found")
		return nil, nil
	}

	module, diags := tfconfig.LoadModule(path)
	if diags.HasErrors() {
		ctx.WithError(diags.Err()).Error("Unable to load Terraform configuration")
		return nil, diags.Err()
	}

	// Get Terraform semantic version for current configuration.
	if len(module.RequiredCore) != 0 {
		version = module.RequiredCore[0]
	} else {
		// No version defined in Terrafom configuration,
		// so we activate the most recent available release.
		ctx.Info("Did not find any version constrainst in Terraform configuration")
		if !Cache.isEmpty() {
			return Cache.LastRelease.Version, nil
		} else {
			ctx.Info("No available Terraform binary")
			return nil, nil
		}
	}

	ctx = log.WithFields(log.Fields{
		"path":    path,
		"version": version,
	})

	v, err := semver.NewVersion(version)

	if err != nil {
		ctx.WithError(err).Error("Unable to extract Terraform version from local configuration")
		return nil, err
	}

	return v, nil
}
