package tfs

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/spf13/viper"
)

// Current software version.
const version = "v0.1.0"

func InitConfig() {
	// Configuration file location is "${XDG_CONFIG_HOME}/tfs"
	// by default, or "${HOME}/.config/tfs" as a fallback.
	if directory, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		viper.AddConfigPath(filepath.Join(directory, "tfs"))
	} else {
		viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "tfs"))
	}

	viper.SetConfigName("config")
	viper.Set("version", version)

	// Set configuration default values.

	// Terraform download URL.
	viper.SetDefault("terraform_download_url", "https://releases.hashicorp.com")

	// File names in the cache will be of the form <prefix> + <semver>.
	viper.SetDefault("terraform_file_name_prefix", "terraform_")

	// Local cache directory is "${XDG_CACHE_HOME}/tfs"
	// by default, or "${HOME}/.cache/tfs" as a fallback.
	if directory, ok := os.LookupEnv("XDG_CACHE_HOME"); ok {
		viper.SetDefault("cache_directory", filepath.Join(directory, "tfs"))
	} else {
		viper.SetDefault("cache_directory", filepath.Join(os.Getenv("HOME"), ".cache", "tfs"))
	}

	// Keep a limited number of release files in the cache.
	viper.SetDefault("cache_auto_clean", true)

	// Number of Terraform releases to keep.
	// Most recent releases will be kept in the cache.
	viper.SetDefault("cache_history", 20) // 20 releases equal roughly 1.2G as of today.

	// Find and read the configuration file.
	err := viper.ReadInConfig()

	ctx := log.WithFields(log.Fields{
		"configDirectory": filepath.Dir(viper.ConfigFileUsed()),
		"fileName":        filepath.Base(viper.ConfigFileUsed()),
	})

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Ignoring this.
		} else {
			ctx.WithError(err).Error("Failed to load configuration")
		}
	} else {
		// Configuration file found and successfully parsed.
		ctx.Info("Configuration loaded")
	}

	// Set cache directory once and for all.
	Cache.Directory = viper.GetString("cache_directory")
}
