package tfs

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Current software version.
const tfsVersion = "v1.4.0"

func InitConfig() {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		slog.Error("Failed to get user home directory", "error", err)
		// This error should really not happen and as far as I'm concerned,
		// is unrecoverable. We have no choice but to give up.
		os.Exit(1)
	}

	// Configuration file location is "${XDG_CONFIG_HOME}/tfs"
	// by default, or "${HOME}/.config/tfs" as a fallback.
	userConfigDir, err := os.UserConfigDir()

	if err != nil {
		userConfigDir = filepath.Join(userHomeDir, ".config")
	}

	viper.AddConfigPath(filepath.Join(userConfigDir, "tfs"))
	viper.SetConfigName("config")

	// Local cache directory is "${XDG_CACHE_HOME}/tfs"
	// by default, or "${HOME}/.cache/tfs" as a fallback.
	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		userCacheDir = filepath.Join(userHomeDir, ".cache")
	}

	/* Configuration default values */

	// Software version.
	viper.Set("version", tfsVersion)

	// User home directory.
	viper.SetDefault("user_home_directory", userHomeDir)

	// User-specific configurations directory.
	viper.SetDefault("user_config_directory", userConfigDir)

	// User-specific executable files directory.
	// TODO: check if the folder belongs to the PATH variable, raise a warning otherwise.
	viper.SetDefault("user_bin_directory", filepath.Join(userHomeDir, ".local", "bin"))

	// Application configuration directory.
	viper.SetDefault("config_directory", filepath.Join(userConfigDir, "tfs"))

	// File names in the cache will be of the form <prefix> + <semver>.
	viper.SetDefault("terraform_file_name_prefix", "terraform_")

	// Application cache directory.
	viper.SetDefault("cache_directory", filepath.Join(userCacheDir, "tfs"))

	// Keep a limited number of release files in the cache.
	viper.SetDefault("cache_auto_clean", true)

	// Number of Terraform releases to keep.
	// Most recent releases will be kept in the cache.
	viper.SetDefault("cache_history", 8)
	viper.SetDefault("cache_minor_version_nb", 0)
	viper.SetDefault("cache_patch_version_nb", 0)

	/* Configuration dynamic values */

	// Find and read the configuration file.
	err = viper.ReadInConfig()

	logger := slog.With(
		"configDirectory", filepath.Dir(viper.ConfigFileUsed()),
		"fileName", filepath.Base(viper.ConfigFileUsed()),
	)

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Ignoring this.
		} else {
			logger.Error("Failed to load tfs configuration", "error", err)
		}
	} else {
		// Configuration file found and successfully parsed.
		if !viper.GetBool("quiet") {
			logger.Info("Configuration loaded")
		}
	}
}
