package tfs

import (
	"testing"

	"github.com/hashicorp/go-version"
)

func mustVersion(t *testing.T, s string) *version.Version {
	t.Helper()
	v, err := version.NewVersion(s)
	if err != nil {
		t.Fatalf("invalid version in test: %v", err)
	}
	return v
}

func makeVersions(t *testing.T, strs []string) []*version.Version {
	t.Helper()
	versions := make([]*version.Version, len(strs))
	for i, s := range strs {
		versions[i] = mustVersion(t, s)
	}
	return versions
}

func TestResolveVersion(t *testing.T) {
	cached := []string{"1.12.0", "1.12.3", "1.12.9", "1.13.0"}

	tests := []struct {
		name        string
		constraint  string
		cached      []string
		expectedVer string
		shouldBeNil bool
		shouldError bool
	}{
		{
			name:        "Pessimistic patch constraint matches highest in range",
			constraint:  "~> 1.12.2",
			cached:      cached,
			expectedVer: "1.12.9",
		},
		{
			name:        "Pessimistic minor constraint",
			constraint:  "~> 1.12",
			cached:      cached,
			expectedVer: "1.13.0", // ~> 1.12 means >= 1.12, < 2.0
		},
		{
			name:        "Greater than or equal",
			constraint:  ">= 1.12.0",
			cached:      cached,
			expectedVer: "1.13.0",
		},
		{
			name:        "Exact version string",
			constraint:  "1.12.3",
			cached:      cached,
			expectedVer: "1.12.3",
		},
		{
			name:        "Exact match with equals operator",
			constraint:  "= 1.12.3",
			cached:      cached,
			expectedVer: "1.12.3",
		},
		{
			name:        "Composite constraint with pessimistic",
			constraint:  "~> 1.12, >= 1.12.3",
			cached:      cached,
			expectedVer: "1.13.0", // ~> 1.12 means >= 1.12, < 2.0
		},
		{
			name:        "No match in empty cache",
			constraint:  "~> 1.12.0",
			cached:      []string{},
			shouldError: true,
		},
		{
			name:        "No match in cache",
			constraint:  "~> 1.14.0",
			cached:      cached,
			shouldError: true,
		},
		{
			name:        "Empty constraint returns nil",
			constraint:  "",
			cached:      cached,
			shouldBeNil: true,
		},
		{
			name:        "Invalid constraint",
			constraint:  "??? bad",
			cached:      cached,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := makeVersions(t, tt.cached)

			result, err := ResolveVersion(tt.constraint, versions)

			if tt.shouldError {
				if err == nil {
					t.Fatalf("expected error for constraint %q, got result %v", tt.constraint, result)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldBeNil {
				if result != nil {
					t.Fatalf("expected nil, got %s", result.String())
				}
				return
			}

			if result == nil {
				t.Fatalf("expected %s, got nil", tt.expectedVer)
			}

			if result.String() != tt.expectedVer {
				t.Fatalf("expected %s, got %s", tt.expectedVer, result.String())
			}
		})
	}
}
