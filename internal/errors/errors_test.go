package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestFileSystemError(t *testing.T) {
	// Test with path
	err1 := NewFileSystemError("read", "/path/to/file", fmt.Errorf("permission denied"))
	expected1 := "read /path/to/file: permission denied"
	if err1.Error() != expected1 {
		t.Errorf("Expected error message '%s', got '%s'", expected1, err1.Error())
	}

	// Test without path
	err2 := NewFileSystemError("parse", "", fmt.Errorf("invalid syntax"))
	expected2 := "parse: invalid syntax"
	if err2.Error() != expected2 {
		t.Errorf("Expected error message '%s', got '%s'", expected2, err2.Error())
	}

	// Test Unwrap
	if err1.Unwrap().Error() != "permission denied" {
		t.Errorf("Unwrap returned incorrect error")
	}

	// Test Is
	testErr := fmt.Errorf("permission denied")
	if err1.Is(testErr) {
		t.Errorf("Is should return false for non-matching error")
	}

	// Test Is with matching error
	permErr := NewFileSystemError("test", "/path", ErrPermissionDenied)
	if !errors.Is(permErr, ErrPermissionDenied) {
		t.Errorf("Is should return true for matching error")
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "FileNotFound",
			err:      ErrFileNotFound,
			expected: true,
		},
		{
			name:     "DirectoryNotFound",
			err:      ErrDirectoryNotFound,
			expected: true,
		},
		{
			name:     "WrappedFileNotFound",
			err:      NewFileSystemError("read", "/path", ErrFileNotFound),
			expected: true,
		},
		{
			name:     "OtherError",
			err:      ErrPermissionDenied,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsPermissionDenied(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "PermissionDenied",
			err:      ErrPermissionDenied,
			expected: true,
		},
		{
			name:     "PathNotAllowed",
			err:      ErrPathNotAllowed,
			expected: true,
		},
		{
			name:     "WrappedPermissionDenied",
			err:      NewFileSystemError("write", "/path", ErrPermissionDenied),
			expected: true,
		},
		{
			name:     "OtherError",
			err:      ErrFileNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPermissionDenied(tt.err)
			if result != tt.expected {
				t.Errorf("IsPermissionDenied(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsInvalidArgument(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "InvalidArgument",
			err:      ErrInvalidArgument,
			expected: true,
		},
		{
			name:     "WrappedInvalidArgument",
			err:      NewFileSystemError("parse", "", ErrInvalidArgument),
			expected: true,
		},
		{
			name:     "OtherError",
			err:      ErrFileNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInvalidArgument(tt.err)
			if result != tt.expected {
				t.Errorf("IsInvalidArgument(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestIsInvalidOperation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "InvalidOperation",
			err:      ErrInvalidOperation,
			expected: true,
		},
		{
			name:     "WrappedInvalidOperation",
			err:      NewFileSystemError("execute", "/path", ErrInvalidOperation),
			expected: true,
		},
		{
			name:     "OtherError",
			err:      ErrFileNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsInvalidOperation(tt.err)
			if result != tt.expected {
				t.Errorf("IsInvalidOperation(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}
