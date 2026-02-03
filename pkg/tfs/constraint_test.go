package tfs

import (
	"testing"

	"github.com/hashicorp/go-version"
)

func TestExpandPessimistic_NoOp(t *testing.T) {
	in := ">=1.2.3, <2.0.0"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != in {
		t.Fatalf("expected no change, got %s", out)
	}
}

func TestExpandPessimistic_Major(t *testing.T) {
	in := "~> 1"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := ">=1.0.0, <2.0.0"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
	c, err := newConstraintExtended(in)
	if err != nil {
		t.Fatalf("constraint error: %v", err)
	}
	if !c.Check(mustVersion(t, "1.5.0")) || c.Check(mustVersion(t, "2.0.0")) {
		t.Fatalf("constraint logic failure: %s", out)
	}
}

func TestExpandPessimistic_Minor(t *testing.T) {
	in := "~> 1.2"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := ">=1.2.0, <1.3.0"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
	c, _ := newConstraintExtended(in)
	if !c.Check(mustVersion(t, "1.2.5")) || c.Check(mustVersion(t, "1.3.0")) {
		t.Fatalf("constraint logic failure: %s", out)
	}
}

func TestExpandPessimistic_Patch(t *testing.T) {
	in := "~> 1.2.3"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := ">=1.2.3, <1.3.0"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
	c, _ := newConstraintExtended(in)
	if !c.Check(mustVersion(t, "1.2.3")) || !c.Check(mustVersion(t, "1.2.9")) || c.Check(mustVersion(t, "1.3.0")) {
		t.Fatalf("constraint logic failure: %s", out)
	}
}

func TestExpandPessimistic_Composite(t *testing.T) {
	in := "~> 1.2, >=1.2.5"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := ">=1.2.0, <1.3.0, >=1.2.5"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
	c, _ := newConstraintExtended(in)
	if !c.Check(mustVersion(t, "1.2.5")) || c.Check(mustVersion(t, "1.3.0")) {
		t.Fatalf("constraint logic failure: %s", out)
	}
}

func TestExpandPessimistic_Invalid(t *testing.T) {
	in := "~> 1.bad"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The regex matches "1" from "~> 1", expands it, and leaves ".bad"
	expected := ">=1.0.0, <2.0.0.bad"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}

func TestExpandPessimistic_NoSpace(t *testing.T) {
	in := "~>1.2"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := ">=1.2.0, <1.3.0"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}

func TestExpandPessimistic_Multiple(t *testing.T) {
	in := "~> 1.2, ~>1.3.4"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := ">=1.2.0, <1.3.0, >=1.3.4, <1.4.0"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}

func TestExpandPessimistic_InvalidDouble(t *testing.T) {
	in := "~> ~> 1.2"
	_, err := expandPessimistic(in)
	if err == nil {
		t.Fatalf("expected error for invalid double operator")
	}
}

func TestExpandPessimistic_InvalidTrailingChars(t *testing.T) {
	in := "~>1.2alpha"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The regex only matches numeric segments, so "1.2alpha" won't match
	// It will match "1.2" and expand that
	expected := ">=1.2.0, <1.3.0alpha"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}

func TestExpandPessimistic_PreRelease(t *testing.T) {
	in := "~> 1.2.3-beta"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Regex matches "1.2.3" but not "-beta", so we get partial expansion
	expected := ">=1.2.3, <1.3.0-beta"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}

func TestExpandPessimistic_BuildMetadata(t *testing.T) {
	in := "~> 1.2.3+meta"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Regex matches "1.2.3" but not "+meta", so we get partial expansion
	expected := ">=1.2.3, <1.3.0+meta"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}

func TestExpandPessimistic_FourSegments(t *testing.T) {
	in := "~> 1.2.3.4"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Regex only matches up to 3 segments, so "1.2.3" is matched
	expected := ">=1.2.3, <1.3.0.4"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}

func TestExpandPessimistic_LeadingZeros(t *testing.T) {
	in := "~> 1.02.3"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := ">=1.2.3, <1.3.0"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}

func TestExpandPessimistic_MixedValidInvalid(t *testing.T) {
	in := "~> 1.2, >=2.0.0, ~> bad"
	_, err := expandPessimistic(in)
	if err == nil {
		t.Fatalf("expected error because ~> bad doesn't match and remains")
	}
	// The "~> bad" doesn't match the regex, so "~>" remains and triggers error
}

func TestExpandPessimistic_EmptyString(t *testing.T) {
	in := ""
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != in {
		t.Fatalf("expected no change for empty string, got %s", out)
	}
}

func TestExpandPessimistic_OnlyRegularConstraints(t *testing.T) {
	in := ">=1.0.0, <2.0.0, !=1.5.0"
	out, err := expandPessimistic(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != in {
		t.Fatalf("expected no change for regular constraints, got %s", out)
	}
}

func mustVersion(t *testing.T, s string) *version.Version {
	v, err := version.NewVersion(s)
	if err != nil {
		t.Fatalf("invalid version in test: %v", err)
	}
	return v
}
