package tfs

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"
)

const (
	tfVersion = "1.3.0"
)

func TestRelease(t *testing.T) {
	t.Run("TestNewRelease", func(t *testing.T) {
		// Initialize test inputs.
		v, _ := version.NewVersion(tfVersion)

		// Call the function being tested.
		output := NewRelease(v)

		// Initialize the expected output.
		expectedOutput := &release{Version: v}

		// Compare the output to the expected output.
		if !reflect.DeepEqual(output, expectedOutput) {
			t.Errorf("NewRelease(%q) = %q, want %q", v, output, expectedOutput)
		}
	})

	t.Run("TestInit", func(t *testing.T) {
		// Initialize test inputs.
		v, _ := version.NewVersion(tfVersion)

		// Call the function being tested.
		output := NewRelease(v).Init()

		// Initialize the expected output.
		filename := viper.GetString("terraform_file_name_prefix") + v.String()
		urlPrefix := fmt.Sprintf("%s/terraform/%s/terraform_%s", viper.GetString("terraform_download_url"), v.String(), v.String())
		binaryURL := fmt.Sprintf("%s_%s_%s.zip", urlPrefix, runtime.GOOS, runtime.GOARCH)
		checksumURL := fmt.Sprintf("%s_SHA256SUMS", urlPrefix)
		url := fmt.Sprintf("%s?checksum=file:%s", binaryURL, checksumURL)

		expectedOutput := &release{
			Version:        v,
			CacheDirectory: Cache.Directory,
			FileName:       filename,
			URLPrefix:      urlPrefix,
			BinaryURL:      binaryURL,
			ChecksumURL:    checksumURL,
			URL:            url,
		}

		// Compare the output to the expected output.
		if !reflect.DeepEqual(output, expectedOutput) {
			t.Errorf("Init(%q) = %q, want %q", v, output, expectedOutput)
		}
	})
}
