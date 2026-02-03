package tfs

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"
)

// versionFromFileName extracts semantic version from Terraform binary name.
func versionFromFileName(fileName string) (*version.Version, error) {
	return version.NewVersion(strings.ReplaceAll(fileName, viper.GetString("terraform_file_name_prefix"), ""))
}

// formatSize returns size in a human readable format.
func formatSize(size uint64) string {
	return humanize.Bytes(size)
}

// expandPessimistic transforms a Terraform/HashiCorp style pessimistic constraint (e.g. "~> 1.2.3")
// into an equivalent go-version constraint expression using >= and < operators.
// Rules:
//
//	~> X      => >= X.0.0, < (X+1).0.0
//	~> X.Y    => >= X.Y.0, < X.(Y+1).0
//	~> X.Y.Z  => >= X.Y.Z, < X.(Y+1).0
//
// Non-numeric or too many segments => error.
var pessimisticConstraintRe = regexp.MustCompile(`~>\s*(\d+(?:\.\d+){0,2})`)

func expandPessimistic(expr string) (string, error) {
	if !strings.Contains(expr, "~>") {
		return expr, nil
	}

	out := expr
	matches := pessimisticConstraintRe.FindAllStringSubmatch(expr, -1)
	for _, m := range matches {
		full := m[0]     // e.g. "~> 1.2.3"
		verToken := m[1] // e.g. "1.2.3"
		segments := strings.Split(verToken, ".")
		if len(segments) == 0 || len(segments) > 3 {
			return expr, fmt.Errorf("invalid pessimistic constraint: %s", expr)
		}
		// parse numeric segments
		ints := make([]int, len(segments))
		for i, s := range segments {
			v, err := strconv.Atoi(s)
			if err != nil {
				return expr, fmt.Errorf("invalid pessimistic constraint: %s", expr)
			}
			ints[i] = v
		}
		major := ints[0]
		var lower, upper string
		switch len(ints) {
		case 1:
			lower = fmt.Sprintf(">=%d.0.0", major)
			upper = fmt.Sprintf("<%d.0.0", major+1)
		case 2:
			minor := ints[1]
			lower = fmt.Sprintf(">=%d.%d.0", major, minor)
			upper = fmt.Sprintf("<%d.%d.0", major, minor+1)
		case 3:
			minor := ints[1]
			patch := ints[2]
			lower = fmt.Sprintf(">=%d.%d.%d", major, minor, patch)
			upper = fmt.Sprintf("<%d.%d.0", major, minor+1)
		}
		replacement := lower + ", " + upper
		out = strings.Replace(out, full, replacement, 1)
	}
	// If any ~> remains, it's invalid (e.g. malformed after first replace or nested operator)
	if strings.Contains(out, "~>") {
		return expr, fmt.Errorf("invalid pessimistic constraint: %s", expr)
	}
	return out, nil
}

// newConstraintExtended wraps version.NewConstraint adding support for the ~> operator.
func newConstraintExtended(expr string) (version.Constraints, error) {
	expanded, err := expandPessimistic(expr)
	if err != nil {
		return nil, err
	}
	return version.NewConstraint(expanded)
}
