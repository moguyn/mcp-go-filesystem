package tools

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/moguyn/mcp-go-filesystem/internal/errors"
	"github.com/moguyn/mcp-go-filesystem/internal/logging"
)

// SearchService implements SearchProvider interface
type SearchService struct {
	allowedDirs []string
	logger      *logging.Logger
	validator   PathValidator
}

// NewSearchService creates a new SearchService
func NewSearchService(allowedDirs []string) *SearchService {
	validator := &PathValidatorImpl{
		allowedDirs: allowedDirs,
	}

	return &SearchService{
		allowedDirs: allowedDirs,
		logger:      logging.DefaultLogger("search_service"),
		validator:   validator,
	}
}

// SearchFiles searches for files matching the query
func (s *SearchService) SearchFiles(query string, path string, recursive bool) ([]SearchResult, error) {
	// Validate path
	validPath, err := s.validator.ValidatePath(path)
	if err != nil {
		return nil, errors.NewFileSystemError("search_files", path, err)
	}

	// Check if the path exists and is a directory
	info, err := os.Stat(validPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewFileSystemError("search_files", path, errors.ErrDirectoryNotFound)
		}
		return nil, errors.NewFileSystemError("search_files", path, err)
	}

	// Perform the search
	results := make([]SearchResult, 0)
	if info.IsDir() {
		err = s.searchInDirectory(validPath, query, recursive, &results)
		if err != nil {
			return nil, errors.NewFileSystemError("search_files", path, err)
		}
	} else {
		err = s.searchInFile(validPath, query, &results)
		if err != nil {
			return nil, errors.NewFileSystemError("search_files", path, err)
		}
	}

	return results, nil
}

// searchInDirectory searches for files in a directory
func (s *SearchService) searchInDirectory(dirPath, query string, recursive bool, results *[]SearchResult) error {
	// Read the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Process each entry
	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())

		// If it's a directory and recursive is true, search in the subdirectory
		if entry.IsDir() && recursive {
			if err := s.searchInDirectory(entryPath, query, recursive, results); err != nil {
				s.logger.Warn("Error searching in directory %s: %v", entryPath, err)
			}
			continue
		}

		// If it's a file, search in the file
		if !entry.IsDir() {
			if err := s.searchInFile(entryPath, query, results); err != nil {
				s.logger.Warn("Error searching in file %s: %v", entryPath, err)
			}
		}
	}

	return nil
}

// searchInFile searches for a query in a file
func (s *SearchService) searchInFile(filePath, query string, results *[]SearchResult) error {
	// Open the file
	file, err := os.Open(filePath) // #nosec G304 - path is validated by ValidatePath in the calling function
	if err != nil {
		return err
	}
	defer file.Close()

	// Check if it's a binary file by reading the first few bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return err
	}

	// Reset file pointer to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	// Check if file appears to be binary
	if isBinary(buf[:n]) {
		s.logger.Debug("Skipping binary file: %s", filePath)
		return nil
	}

	// Read the file line by line
	scanner := bufio.NewScanner(file)

	// Increase buffer size to handle longer lines
	const maxScanTokenSize = 1024 * 1024 * 10 // 10MB buffer
	scanBuf := make([]byte, maxScanTokenSize)
	scanner.Buffer(scanBuf, maxScanTokenSize)

	lineNum := 1
	query = strings.ToLower(query)
	for scanner.Scan() {
		content := scanner.Text()
		line := strings.ToLower(content)
		if strings.Contains(line, query) {
			// Add the result
			*results = append(*results, SearchResult{
				Path:    filePath,
				Line:    lineNum,
				Content: content,
			})
		}
		lineNum++
	}

	return scanner.Err()
}

// isBinary checks if data appears to be binary content
func isBinary(data []byte) bool {
	// Check for null bytes, which are common in binary files
	for _, b := range data {
		if b == 0 {
			return true
		}
	}

	// Count control characters (except common ones like tab, newline)
	controlCount := 0
	for _, b := range data {
		if (b < 32 && b != 9 && b != 10 && b != 13) || b >= 127 {
			controlCount++
		}
	}

	// If more than 10% are control characters, consider it binary
	return controlCount > len(data)/10
}
