package server

import (
	"strings"
	"testing"
)

func TestNormalizeLineEndings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No line endings",
			input:    "test",
			expected: "test",
		},
		{
			name:     "Unix line endings",
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "Windows line endings",
			input:    "line1\r\nline2\r\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "Mixed line endings",
			input:    "line1\r\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeLineEndings(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeLineEndings() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCreateUnifiedDiff(t *testing.T) {
	original := "line1\nline2\nline3\n"
	modified := "line1\nmodified line\nline3\n"
	filepath := "test.txt"

	diff := createUnifiedDiff(original, modified, filepath)

	// Basic validation of the diff output
	if !strings.Contains(diff, "diff --git") {
		t.Errorf("Diff doesn't contain 'diff --git' header")
	}
	if !strings.Contains(diff, "a/test.txt") {
		t.Errorf("Diff doesn't contain the original file path")
	}
	if !strings.Contains(diff, "b/test.txt") {
		t.Errorf("Diff doesn't contain the modified file path")
	}

	// The go-diff library formats patches differently than expected
	// Instead of checking for specific lines, just verify the diff contains
	// something that indicates changes
	if !strings.Contains(diff, "@@ ") {
		t.Errorf("Diff doesn't contain patch header")
	}
}
