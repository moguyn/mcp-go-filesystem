package tools

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/moguyn/mcp-go-filesystem/internal/errors"
	"github.com/moguyn/mcp-go-filesystem/internal/logging"
)

// FileService implements FileReader, FileWriter, and FileManager interfaces
type FileService struct {
	allowedDirs []string
	logger      *logging.Logger
	validator   PathValidator
}

// NewFileService creates a new FileService
func NewFileService(allowedDirs []string) *FileService {
	validator := &PathValidatorImpl{
		allowedDirs: allowedDirs,
	}

	return &FileService{
		allowedDirs: allowedDirs,
		logger:      logging.DefaultLogger("file_service"),
		validator:   validator,
	}
}

// ReadFile reads the content of a file
func (s *FileService) ReadFile(path string) (string, error) {
	// Validate path
	validPath, err := s.validator.ValidatePath(path)
	if err != nil {
		return "", errors.NewFileSystemError("read_file", path, err)
	}

	// Read file
	content, err := os.ReadFile(validPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.NewFileSystemError("read_file", path, errors.ErrFileNotFound)
		}
		return "", errors.NewFileSystemError("read_file", path, err)
	}

	return string(content), nil
}

// ReadMultipleFiles reads the content of multiple files
func (s *FileService) ReadMultipleFiles(paths []string) ([]FileContent, error) {
	results := make([]FileContent, 0, len(paths))

	for _, path := range paths {
		content, err := s.ReadFile(path)
		fileContent := FileContent{
			Path: path,
		}

		if err != nil {
			fileContent.Error = err.Error()
		} else {
			fileContent.Content = content
		}

		results = append(results, fileContent)
	}

	return results, nil
}

// WriteFile writes content to a file
func (s *FileService) WriteFile(path, content string, append bool) error {
	// Validate path
	validPath, err := s.validator.ValidatePath(path)
	if err != nil {
		return errors.NewFileSystemError("write_file", path, err)
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewFileSystemError("write_file", path, err)
	}

	// Open file with appropriate flags
	flag := os.O_WRONLY | os.O_CREATE
	if append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	file, err := os.OpenFile(validPath, flag, 0644)
	if err != nil {
		return errors.NewFileSystemError("write_file", path, err)
	}
	defer file.Close()

	// Write content
	if _, err := file.WriteString(content); err != nil {
		return errors.NewFileSystemError("write_file", path, err)
	}

	return nil
}

// EditFile edits a portion of a file
func (s *FileService) EditFile(path, content string, startLine, endLine int) error {
	// Validate path
	validPath, err := s.validator.ValidatePath(path)
	if err != nil {
		return errors.NewFileSystemError("edit_file", path, err)
	}

	// Read the entire file
	fileBytes, err := os.ReadFile(validPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewFileSystemError("edit_file", path, errors.ErrFileNotFound)
		}
		return errors.NewFileSystemError("edit_file", path, err)
	}
	fileContent := string(fileBytes)

	// Split the file into lines
	lines := strings.Split(fileContent, "\n")

	// Validate line numbers
	if startLine < 1 || startLine > len(lines) {
		return errors.NewFileSystemError("edit_file", path, errors.ErrInvalidArgument)
	}
	if endLine < startLine || endLine > len(lines) {
		return errors.NewFileSystemError("edit_file", path, errors.ErrInvalidArgument)
	}

	// Replace the specified lines
	newLines := strings.Split(content, "\n")
	lines = append(append(lines[:startLine-1], newLines...), lines[endLine:]...)

	// Join the lines back together
	newContent := strings.Join(lines, "\n")

	// Write the file
	return s.WriteFile(path, newContent, false)
}

// DeleteFile deletes a file
func (s *FileService) DeleteFile(path string) error {
	// Validate path
	validPath, err := s.validator.ValidatePath(path)
	if err != nil {
		return errors.NewFileSystemError("delete_file", path, err)
	}

	// Check if the path exists and is a file
	info, err := os.Stat(validPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewFileSystemError("delete_file", path, errors.ErrFileNotFound)
		}
		return errors.NewFileSystemError("delete_file", path, err)
	}
	if info.IsDir() {
		return errors.NewFileSystemError("delete_file", path, errors.ErrInvalidOperation)
	}

	// Delete the file
	if err := os.Remove(validPath); err != nil {
		return errors.NewFileSystemError("delete_file", path, err)
	}

	return nil
}

// MoveFile moves a file from one location to another
func (s *FileService) MoveFile(sourcePath, destinationPath string) error {
	// Validate source path
	validSourcePath, err := s.validator.ValidatePath(sourcePath)
	if err != nil {
		return errors.NewFileSystemError("move_file", sourcePath, err)
	}

	// Validate destination path
	validDestPath, err := s.validator.ValidatePath(destinationPath)
	if err != nil {
		return errors.NewFileSystemError("move_file", destinationPath, err)
	}

	// Check if the source exists and is a file
	info, err := os.Stat(validSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewFileSystemError("move_file", sourcePath, errors.ErrFileNotFound)
		}
		return errors.NewFileSystemError("move_file", sourcePath, err)
	}
	if info.IsDir() {
		return errors.NewFileSystemError("move_file", sourcePath, errors.ErrInvalidOperation)
	}

	// Create parent directories if they don't exist
	destDir := filepath.Dir(validDestPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.NewFileSystemError("move_file", destinationPath, err)
	}

	// Move the file
	if err := os.Rename(validSourcePath, validDestPath); err != nil {
		return errors.NewFileSystemError("move_file", sourcePath, err)
	}

	return nil
}

// CopyFile copies a file from one location to another
func (s *FileService) CopyFile(sourcePath, destinationPath string) error {
	// Validate source path
	validSourcePath, err := s.validator.ValidatePath(sourcePath)
	if err != nil {
		return errors.NewFileSystemError("copy_file", sourcePath, err)
	}

	// Validate destination path
	validDestPath, err := s.validator.ValidatePath(destinationPath)
	if err != nil {
		return errors.NewFileSystemError("copy_file", destinationPath, err)
	}

	// Check if the source exists and is a file
	info, err := os.Stat(validSourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewFileSystemError("copy_file", sourcePath, errors.ErrFileNotFound)
		}
		return errors.NewFileSystemError("copy_file", sourcePath, err)
	}
	if info.IsDir() {
		return errors.NewFileSystemError("copy_file", sourcePath, errors.ErrInvalidOperation)
	}

	// Create parent directories if they don't exist
	destDir := filepath.Dir(validDestPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.NewFileSystemError("copy_file", destinationPath, err)
	}

	// Open source file
	source, err := os.Open(validSourcePath)
	if err != nil {
		return errors.NewFileSystemError("copy_file", sourcePath, err)
	}
	defer source.Close()

	// Create destination file
	destination, err := os.Create(validDestPath)
	if err != nil {
		return errors.NewFileSystemError("copy_file", destinationPath, err)
	}
	defer destination.Close()

	// Copy the file
	if _, err := io.Copy(destination, source); err != nil {
		return errors.NewFileSystemError("copy_file", sourcePath, err)
	}

	// Preserve file mode
	if err := os.Chmod(validDestPath, info.Mode()); err != nil {
		return errors.NewFileSystemError("copy_file", destinationPath, err)
	}

	return nil
}

// PathValidatorImpl implements PathValidator interface
type PathValidatorImpl struct {
	allowedDirs []string
}

// ValidatePath validates that a path is within the allowed directories
func (v *PathValidatorImpl) ValidatePath(requestedPath string) (string, error) {
	// Check for invalid characters in the path
	if strings.ContainsRune(requestedPath, 0) {
		return "", errors.ErrInvalidPath
	}

	// Normalize the path
	expandedPath := ExpandHome(requestedPath)
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", err
	}
	normalizedPath := filepath.Clean(absPath)

	// Check if the path is within any of the allowed directories
	for _, allowedDir := range v.allowedDirs {
		// Normalize the allowed directory
		allowedDirAbs, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}
		allowedDirNormalized := filepath.Clean(allowedDirAbs)

		// Check if the path is the allowed directory or a subdirectory
		if normalizedPath == allowedDirNormalized || strings.HasPrefix(normalizedPath, allowedDirNormalized+string(filepath.Separator)) {
			return normalizedPath, nil
		}
	}

	return "", errors.ErrPathNotAllowed
}
