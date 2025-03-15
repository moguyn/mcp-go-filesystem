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
	if !info.IsDir() {
		return nil, errors.NewFileSystemError("search_files", path, errors.ErrInvalidOperation)
	}

	// Perform the search
	results := make([]SearchResult, 0)
	err = s.searchInDirectory(validPath, query, recursive, &results)
	if err != nil {
		return nil, errors.NewFileSystemError("search_files", path, err)
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

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, query) {
			// Add the result
			*results = append(*results, SearchResult{
				Path:    filePath,
				Line:    lineNum,
				Content: line,
			})
		}
		lineNum++
	}

	return scanner.Err()
}
