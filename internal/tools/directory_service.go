package tools

import (
	"os"
	"path/filepath"
	"time"

	"github.com/moguyn/mcp-go-filesystem/internal/errors"
	"github.com/moguyn/mcp-go-filesystem/internal/logging"
)

// DirectoryService implements DirectoryManager interface
type DirectoryService struct {
	allowedDirs []string
	logger      *logging.Logger
	validator   PathValidator
}

// NewDirectoryService creates a new DirectoryService
func NewDirectoryService(allowedDirs []string) *DirectoryService {
	validator := &PathValidatorImpl{
		allowedDirs: allowedDirs,
	}

	return &DirectoryService{
		allowedDirs: allowedDirs,
		logger:      logging.DefaultLogger("directory_service"),
		validator:   validator,
	}
}

// CreateDirectory creates a new directory
func (s *DirectoryService) CreateDirectory(path string) error {
	// Validate path
	validPath, err := s.validator.ValidatePath(path)
	if err != nil {
		return errors.NewFileSystemError("create_directory", path, err)
	}

	// Create the directory
	if err := os.MkdirAll(validPath, 0750); err != nil {
		return errors.NewFileSystemError("create_directory", path, err)
	}

	return nil
}

// ListDirectory lists the contents of a directory
func (s *DirectoryService) ListDirectory(path string) ([]FileInfo, error) {
	// Validate path
	validPath, err := s.validator.ValidatePath(path)
	if err != nil {
		return nil, errors.NewFileSystemError("list_directory", path, err)
	}

	// Check if the path exists and is a directory
	info, err := os.Stat(validPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewFileSystemError("list_directory", path, errors.ErrDirectoryNotFound)
		}
		return nil, errors.NewFileSystemError("list_directory", path, err)
	}
	if !info.IsDir() {
		return nil, errors.NewFileSystemError("list_directory", path, errors.ErrInvalidOperation)
	}

	// Read the directory
	entries, err := os.ReadDir(validPath)
	if err != nil {
		return nil, errors.NewFileSystemError("list_directory", path, err)
	}

	// Convert entries to FileInfo
	result := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		entryInfo, err := entry.Info()
		if err != nil {
			s.logger.Warn("Error getting info for %s: %v", entry.Name(), err)
			continue
		}

		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(path, entry.Name()),
			Size:    entryInfo.Size(),
			IsDir:   entry.IsDir(),
			ModTime: entryInfo.ModTime().Format(time.RFC3339),
		}

		if !entry.IsDir() {
			fileInfo.Extension = filepath.Ext(entry.Name())
		}

		result = append(result, fileInfo)
	}

	return result, nil
}

// DeleteDirectory deletes a directory
func (s *DirectoryService) DeleteDirectory(path string, recursive bool) error {
	// Validate path
	validPath, err := s.validator.ValidatePath(path)
	if err != nil {
		return errors.NewFileSystemError("delete_directory", path, err)
	}

	// Check if the path exists and is a directory
	info, err := os.Stat(validPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewFileSystemError("delete_directory", path, errors.ErrDirectoryNotFound)
		}
		return errors.NewFileSystemError("delete_directory", path, err)
	}
	if !info.IsDir() {
		return errors.NewFileSystemError("delete_directory", path, errors.ErrInvalidOperation)
	}

	// If not recursive, check if the directory is empty
	if !recursive {
		entries, err := os.ReadDir(validPath)
		if err != nil {
			return errors.NewFileSystemError("delete_directory", path, err)
		}
		if len(entries) > 0 {
			return errors.NewFileSystemError("delete_directory", path, errors.ErrInvalidOperation)
		}

		// Delete the empty directory
		if err := os.Remove(validPath); err != nil {
			return errors.NewFileSystemError("delete_directory", path, err)
		}
	} else {
		// Delete the directory and all its contents
		if err := os.RemoveAll(validPath); err != nil {
			return errors.NewFileSystemError("delete_directory", path, err)
		}
	}

	return nil
}
