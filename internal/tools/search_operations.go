package tools

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleSearchFiles handles the search_files tool
func handleSearchFiles(request mcp.CallToolRequest, allowedDirectories []string) (*mcp.CallToolResult, error) {
	root, ok := request.Params.Arguments["root"].(string)
	if !ok {
		return nil, fmt.Errorf("root must be a string")
	}

	pattern, ok := request.Params.Arguments["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("pattern must be a string")
	}

	// Parse exclude patterns
	var excludePatterns []string
	if excludeJSON, ok := request.Params.Arguments["exclude"].(string); ok && excludeJSON != "" {
		if err := json.Unmarshal([]byte(excludeJSON), &excludePatterns); err != nil {
			return nil, fmt.Errorf("invalid exclude JSON array: %v", err)
		}
	}

	// Validate path
	validRoot, err := ValidatePath(root, allowedDirectories)
	if err != nil {
		return nil, fmt.Errorf("invalid root path: %v", err)
	}

	// Search files
	matches, err := searchFiles(validRoot, pattern, excludePatterns)
	if err != nil {
		return nil, fmt.Errorf("error searching files: %v", err)
	}

	// Format results
	var resultBuilder strings.Builder
	resultBuilder.WriteString(fmt.Sprintf("Found %d matches for pattern '%s' in %s:\n\n", len(matches), pattern, root))
	for _, match := range matches {
		resultBuilder.WriteString(fmt.Sprintf("%s\n", match))
	}

	result := &mcp.CallToolResult{
		Content: []interface{}{
			&mcp.TextContent{
				Type: "text",
				Text: resultBuilder.String(),
			},
		},
	}
	return result, nil
}

// searchFiles searches for files matching a pattern
func searchFiles(rootPath, pattern string, excludePatterns []string) ([]string, error) {
	var matches []string

	// Compile exclude patterns
	var excludeRegexps []*regexp.Regexp
	for _, excludePattern := range excludePatterns {
		// Convert glob pattern to regexp
		regexpPattern := "^" + strings.ReplaceAll(strings.ReplaceAll(excludePattern, ".", "\\."), "*", ".*") + "$"
		re, err := regexp.Compile(regexpPattern)
		if err != nil {
			continue
		}
		excludeRegexps = append(excludeRegexps, re)
	}

	// Walk directory
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if file matches pattern
		match, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return nil // Skip invalid patterns
		}

		if match {
			// Check if file is excluded
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return nil
			}

			excluded := false
			for _, re := range excludeRegexps {
				if re.MatchString(relPath) || re.MatchString(filepath.Base(path)) {
					excluded = true
					break
				}
			}

			if !excluded {
				matches = append(matches, path)
			}
		}

		return nil
	})

	return matches, err
}
