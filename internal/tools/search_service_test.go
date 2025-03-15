package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSearchService(t *testing.T) {
	// Test creating a new search service
	allowedDirs := []string{"/tmp", "/var"}
	service := NewSearchService(allowedDirs)

	assert.NotNil(t, service)
	assert.Equal(t, allowedDirs, service.allowedDirs)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.validator)
}

func TestSearchService_SearchFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-search-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectories
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files with content
	files := map[string]string{
		filepath.Join(tmpDir, "file1.txt"):       "This is a test file with test content",
		filepath.Join(tmpDir, "file2.txt"):       "Another file without the search term",
		filepath.Join(subDir, "nested-file.txt"): "Nested file with test content",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Initialize service with allowed directories
	service := NewSearchService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name          string
		query         string
		path          string
		recursive     bool
		expectError   bool
		expectedCount int
	}{
		{
			name:          "Search in root directory without recursion",
			query:         "test",
			path:          tmpDir,
			recursive:     false,
			expectError:   false,
			expectedCount: 1, // Only file1.txt in root contains "test"
		},
		{
			name:          "Search in root directory with recursion",
			query:         "test",
			path:          tmpDir,
			recursive:     true,
			expectError:   false,
			expectedCount: 2, // file1.txt in root and nested-file.txt contain "test"
		},
		{
			name:          "Search in subdirectory",
			query:         "Nested",
			path:          subDir,
			recursive:     false,
			expectError:   false,
			expectedCount: 1, // Only nested-file.txt contains "Nested"
		},
		{
			name:          "Search with no matches",
			query:         "nonexistent",
			path:          tmpDir,
			recursive:     true,
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:          "Search in non-existent directory",
			query:         "test",
			path:          filepath.Join(tmpDir, "non-existent"),
			recursive:     false,
			expectError:   true,
			expectedCount: 0,
		},
		{
			name:          "Search outside allowed path",
			query:         "test",
			path:          os.TempDir(),
			recursive:     false,
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results, err := service.SearchFiles(tc.query, tc.path, tc.recursive)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, len(results))

				// Verify search results
				for _, result := range results {
					// Check that the result contains the query
					assert.Contains(t, result.Content, tc.query)

					// Check that the path is within the search directory
					assert.True(t, filepath.IsAbs(result.Path))

					// If not recursive, ensure the file is directly in the search directory
					if !tc.recursive && tc.path != subDir {
						assert.Equal(t, filepath.Dir(result.Path), tmpDir)
					}
				}
			}
		})
	}
}

func TestSearchService_searchInDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-search-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectories
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test files with content
	files := map[string]string{
		filepath.Join(tmpDir, "file1.txt"):       "This is a test file with test content",
		filepath.Join(tmpDir, "file2.txt"):       "Another file without the search term",
		filepath.Join(subDir, "nested-file.txt"): "Nested file with test content",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Initialize service with allowed directories
	service := NewSearchService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name          string
		query         string
		path          string
		recursive     bool
		expectedCount int
	}{
		{
			name:          "Search in directory without recursion",
			query:         "test",
			path:          tmpDir,
			recursive:     false,
			expectedCount: 1, // Only file1.txt in root contains "test"
		},
		{
			name:          "Search in directory with recursion",
			query:         "test",
			path:          tmpDir,
			recursive:     true,
			expectedCount: 2, // file1.txt in root and nested-file.txt contain "test"
		},
		{
			name:          "Search with no matches",
			query:         "nonexistent",
			path:          tmpDir,
			recursive:     true,
			expectedCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results := make([]SearchResult, 0)
			err := service.searchInDirectory(tc.path, tc.query, tc.recursive, &results)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCount, len(results))

			// Verify search results
			for _, result := range results {
				// Check that the result contains the query
				assert.Contains(t, result.Content, tc.query)
			}
		})
	}
}

func TestSearchService_searchInFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-search-service-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files with content
	multilineContent := "Line 1: no match\nLine 2: test match\nLine 3: another test\nLine 4: no match"
	multilineFile := filepath.Join(tmpDir, "multiline.txt")
	if err := os.WriteFile(multilineFile, []byte(multilineContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize service with allowed directories
	service := NewSearchService([]string{tmpDir})

	// Test cases
	tests := []struct {
		name          string
		query         string
		path          string
		expectedCount int
		expectedLines []int
	}{
		{
			name:          "Search with multiple matches",
			query:         "test",
			path:          multilineFile,
			expectedCount: 2,
			expectedLines: []int{2, 3}, // Line 2 and Line 3 contain "test"
		},
		{
			name:          "Search with single match",
			query:         "another",
			path:          multilineFile,
			expectedCount: 1,
			expectedLines: []int{3}, // Only Line 3 contains "another"
		},
		{
			name:          "Search with no matches",
			query:         "nonexistent",
			path:          multilineFile,
			expectedCount: 0,
			expectedLines: []int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results := make([]SearchResult, 0)
			err := service.searchInFile(tc.path, tc.query, &results)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCount, len(results))

			// Verify search results
			foundLines := make([]int, 0, len(results))
			for _, result := range results {
				// Check that the result contains the query
				assert.Contains(t, result.Content, tc.query)
				assert.Equal(t, tc.path, result.Path)
				foundLines = append(foundLines, result.Line)
			}

			// Check that we found matches on the expected lines
			for _, expectedLine := range tc.expectedLines {
				found := false
				for _, foundLine := range foundLines {
					if foundLine == expectedLine {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find a match on line %d", expectedLine)
			}
		})
	}
}
