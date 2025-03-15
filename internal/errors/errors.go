package errors

import (
	"errors"
	"fmt"
)

// Standard error types
var (
	ErrInvalidPath       = errors.New("invalid path")
	ErrPathNotAllowed    = errors.New("path not within allowed directories")
	ErrFileNotFound      = errors.New("file not found")
	ErrDirectoryNotFound = errors.New("directory not found")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrInvalidOperation  = errors.New("invalid operation")
)

// FileSystemError represents an error related to filesystem operations
type FileSystemError struct {
	Op   string // Operation that failed
	Path string // Path related to the error
	Err  error  // Underlying error
}

// Error returns the error message
func (e *FileSystemError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
}

// Unwrap returns the underlying error
func (e *FileSystemError) Unwrap() error {
	return e.Err
}

// Is reports whether target is in err's chain
func (e *FileSystemError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewFileSystemError creates a new FileSystemError
func NewFileSystemError(op, path string, err error) *FileSystemError {
	return &FileSystemError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// IsNotFound returns true if the error indicates a not found condition
func IsNotFound(err error) bool {
	return errors.Is(err, ErrFileNotFound) || errors.Is(err, ErrDirectoryNotFound)
}

// IsPermissionDenied returns true if the error indicates a permission denied condition
func IsPermissionDenied(err error) bool {
	return errors.Is(err, ErrPermissionDenied) || errors.Is(err, ErrPathNotAllowed)
}

// IsInvalidArgument returns true if the error indicates an invalid argument
func IsInvalidArgument(err error) bool {
	return errors.Is(err, ErrInvalidArgument)
}

// IsInvalidOperation returns true if the error indicates an invalid operation
func IsInvalidOperation(err error) bool {
	return errors.Is(err, ErrInvalidOperation)
}
