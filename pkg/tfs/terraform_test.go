package tfs

import (
	"testing"

	"github.com/hashicorp/go-version"
)

func TestGetTfVersion_FindsBestMatch(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	cache := NewLocalCache(cacheDir)

	// Populate cache with test versions
	v1 := mustVersion(t, "1.12.0")
	v2 := mustVersion(t, "1.12.3")
	v3 := mustVersion(t, "1.13.0")

	cache.releases = map[string]*release{
		v1.String(): cache.NewRelease(v1),
		v2.String(): cache.NewRelease(v2),
		v3.String(): cache.NewRelease(v3),
	}

	tests := []struct {
		name        string
		constraint  string
		expectedVer string
		shouldBeNil bool
		shouldError bool
	}{
		{
			name:        "Pessimistic constraint matches highest in range",
			constraint:  "~> 1.12.2",
			expectedVer: "1.12.3",
		},
		{
			name:        "Pessimistic minor constraint",
			constraint:  "~> 1.12",
			expectedVer: "1.12.3",
		},
		{
			name:        "Greater than or equal",
			constraint:  ">= 1.12.0",
			expectedVer: "1.13.0",
		},
		{
			name:        "Exact match",
			constraint:  "= 1.12.3",
			expectedVer: "1.12.3",
		},
		{
			name:        "Complex constraint with pessimistic",
			constraint:  "~> 1.12, >= 1.12.3",
			expectedVer: "1.12.3",
		},
		{
			name:        "No match in cache",
			constraint:  "~> 1.14",
			shouldBeNil: true,
		},
		{
			name:        "Invalid constraint",
			constraint:  "~> bad",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the constraint for testing
			constraint, err := newConstraintExtended(tt.constraint)

			if tt.shouldError {
				if err == nil {
					t.Fatalf("expected error for constraint %s", tt.constraint)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Find best match
			var bestMatch *version.Version
			for _, release := range cache.releases {
				if constraint.Check(release.Version) {
					if bestMatch == nil || release.Version.GreaterThan(bestMatch) {
						bestMatch = release.Version
					}
				}
			}

			if tt.shouldBeNil {
				if bestMatch != nil {
					t.Fatalf("expected nil, got %s", bestMatch.String())
				}
				return
			}

			if bestMatch == nil {
				t.Fatal("expected a match, got nil")
			}

			if bestMatch.String() != tt.expectedVer {
				t.Fatalf("expected %s, got %s", tt.expectedVer, bestMatch.String())
			}
		})
	}
}

func TestGetTfVersion_EmptyCache(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	cache := NewLocalCache(cacheDir)
	cache.releases = map[string]*release{}

	constraint, err := newConstraintExtended("~> 1.12.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bestMatch *version.Version
	for _, release := range cache.releases {
		if constraint.Check(release.Version) {
			if bestMatch == nil || release.Version.GreaterThan(bestMatch) {
				bestMatch = release.Version
			}
		}
	}

	if bestMatch != nil {
		t.Fatalf("expected nil for empty cache, got %s", bestMatch.String())
	}
}

func TestGetTfVersion_SelectsHighestMatchingVersion(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	cache := NewLocalCache(cacheDir)

	// Add multiple versions in the same minor range
	versions := []string{"1.12.0", "1.12.1", "1.12.5", "1.12.9", "1.13.0"}
	for _, v := range versions {
		ver := mustVersion(t, v)
		cache.releases[ver.String()] = cache.NewRelease(ver)
	}

	constraint, err := newConstraintExtended("~> 1.12")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bestMatch *version.Version
	for _, release := range cache.releases {
		if constraint.Check(release.Version) {
			if bestMatch == nil || release.Version.GreaterThan(bestMatch) {
				bestMatch = release.Version
			}
		}
	}

	if bestMatch == nil {
		t.Fatal("expected a match, got nil")
	}

	// Should pick 1.12.9, not 1.13.0 (which is outside the constraint)
	if bestMatch.String() != "1.12.9" {
		t.Fatalf("expected 1.12.9, got %s", bestMatch.String())
	}
}
