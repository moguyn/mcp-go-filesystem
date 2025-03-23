package tools

import (
	"os"
	"path/filepath"
	"strings"
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
			name:          "Search with case insensitive match",
			query:         "TEST",
			path:          multilineFile,
			expectedCount: 2,
			expectedLines: []int{2, 3}, // Line 2 and Line 3 contain "TEST"
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
				assert.Contains(t, strings.ToLower(result.Content), strings.ToLower(tc.query))
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

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "Text file with ASCII content",
			data:     []byte("This is a plain text file with normal content."),
			expected: false,
		},
		{
			name:     "Text file with newlines and tabs",
			data:     []byte("Line 1\nLine 2\tTabbed content"),
			expected: false,
		},
		{
			name:     "Binary file with null bytes",
			data:     []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x00, 0x57, 0x6f, 0x72, 0x6c, 0x64},
			expected: true,
		},
		{
			name:     "Binary file with high control character ratio",
			data:     []byte{0x48, 0x65, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x6c, 0x6f},
			expected: true,
		},
		{
			name:     "Text file with a few control characters",
			data:     []byte("Hello\x01World"), // Just one control character
			expected: false,
		},
		{
			name:     "Empty data",
			data:     []byte{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isBinary(tc.data)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSearchService_searchInFile_BinaryDetection(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-search-binary-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a text file
	textFile := filepath.Join(tmpDir, "text.txt")
	textContent := "This is a text file with the word 'binary' in it"
	if err := os.WriteFile(textFile, []byte(textContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a binary file
	binaryFile := filepath.Join(tmpDir, "binary.bin")
	binaryContent := []byte{0x7F, 0x45, 0x4C, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	binaryContent = append(binaryContent, []byte("contains the search term binary")...)
	if err := os.WriteFile(binaryFile, binaryContent, 0644); err != nil {
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
		isBinary      bool
	}{
		{
			name:          "Search in text file",
			query:         "binary",
			path:          textFile,
			expectedCount: 1,
			isBinary:      false,
		},
		{
			name:          "Search in binary file",
			query:         "binary",
			path:          binaryFile,
			expectedCount: 0, // Should be skipped due to binary detection
			isBinary:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results := make([]SearchResult, 0)
			err := service.searchInFile(tc.path, tc.query, &results)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedCount, len(results))

			// For text files, verify the content contains the query
			if !tc.isBinary && len(results) > 0 {
				for _, result := range results {
					assert.Contains(t, result.Content, tc.query)
				}
			}
		})
	}
}

func TestSearchService_searchInFile_LongLines(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-search-longlines-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with a very long line
	longLineFile := filepath.Join(tmpDir, "longline.txt")

	// Create a line that's longer than the default scanner buffer (64KB)
	// but shorter than our increased buffer size (1MB)
	longLine := strings.Repeat("a", 100000) + "FINDME" + strings.Repeat("b", 100000)

	if err := os.WriteFile(longLineFile, []byte(longLine), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize service with allowed directories
	service := NewSearchService([]string{tmpDir})

	// Test searching in the file with a long line
	results := make([]SearchResult, 0)
	err = service.searchInFile(longLineFile, "FINDME", &results)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))

	if len(results) > 0 {
		assert.Contains(t, results[0].Content, "FINDME")
	}
}
