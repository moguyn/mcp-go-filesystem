package server

import (
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// normalizeLineEndings ensures consistent line endings
func normalizeLineEndings(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}

// createUnifiedDiff creates a git-style unified diff
func createUnifiedDiff(originalContent, newContent, filepath string) string {
	// Ensure consistent line endings
	originalContent = normalizeLineEndings(originalContent)
	newContent = normalizeLineEndings(newContent)

	// Use diffmatchpatch to create a diff
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(originalContent, newContent, false)

	// Format as unified diff
	patches := dmp.PatchMake(originalContent, diffs)
	patchText := dmp.PatchToText(patches)

	// Format the diff with appropriate number of backticks
	numBackticks := 3
	for strings.Contains(patchText, strings.Repeat("`", numBackticks)) {
		numBackticks++
	}

	return fmt.Sprintf("%s\ndiff --git a/%s b/%s\n--- a/%s\n+++ b/%s\n%s\n%s\n\n",
		strings.Repeat("`", numBackticks),
		filepath, filepath,
		filepath, filepath,
		patchText,
		strings.Repeat("`", numBackticks))
}
