package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	version := "1.0.0"
	cfg := DefaultConfig(version)

	if cfg.Version != version {
		t.Errorf("Expected version %s, got %s", version, cfg.Version)
	}
	if cfg.ServerMode != StdioMode {
		t.Errorf("Expected server mode %s, got %s", StdioMode, cfg.ServerMode)
	}
	if cfg.ListenAddr != "0.0.0.0:38085" {
		t.Errorf("Expected listen address 0.0.0.0:38085, got %s", cfg.ListenAddr)
	}
	if len(cfg.AllowedDirs) != 0 {
		t.Errorf("Expected empty allowed dirs, got %v", cfg.AllowedDirs)
	}
	if cfg.LogLevel != "INFO" {
		t.Errorf("Expected log level INFO, got %s", cfg.LogLevel)
	}
}

func TestParseCommandLineArgs(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		checkConfig func(*Config) bool
	}{
		{
			name:        "Help flag",
			args:        []string{"cmd", "--help"},
			expectError: true,
		},
		{
			name:        "No arguments",
			args:        []string{"cmd"},
			expectError: true,
		},
		{
			name:        "Invalid mode",
			args:        []string{"cmd", "--mode=invalid", tempDir},
			expectError: true,
		},
		{
			name:        "Valid stdio mode",
			args:        []string{"cmd", "--mode=stdio", tempDir},
			expectError: false,
			checkConfig: func(cfg *Config) bool {
				return cfg.ServerMode == StdioMode && len(cfg.AllowedDirs) == 1
			},
		},
		{
			name:        "Valid sse mode with custom listen address",
			args:        []string{"cmd", "--mode=sse", "--listen=127.0.0.1:8080", tempDir},
			expectError: false,
			checkConfig: func(cfg *Config) bool {
				return cfg.ServerMode == SSEMode && cfg.ListenAddr == "127.0.0.1:8080" && len(cfg.AllowedDirs) == 1
			},
		},
		{
			name:        "Custom log level",
			args:        []string{"cmd", "--log-level=debug", tempDir},
			expectError: false,
			checkConfig: func(cfg *Config) bool {
				return cfg.LogLevel == "DEBUG" && len(cfg.AllowedDirs) == 1
			},
		},
		{
			name:        "Multiple directories",
			args:        []string{"cmd", tempDir, tempDir},
			expectError: false,
			checkConfig: func(cfg *Config) bool {
				return len(cfg.AllowedDirs) == 2
			},
		},
		{
			name:        "Invalid directory",
			args:        []string{"cmd", "/path/that/does/not/exist"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseCommandLineArgs("1.0.0", tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if tt.checkConfig != nil && !tt.checkConfig(cfg) {
					t.Errorf("Config validation failed")
				}
			}
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "validate-dir-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "validate-dir-test-file")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Valid directory",
			path:        tempDir,
			expectError: false,
		},
		{
			name:        "Non-existent directory",
			path:        "/path/that/does/not/exist",
			expectError: true,
		},
		{
			name:        "File instead of directory",
			path:        tempFile.Name(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := validateDirectory(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Check that the path is absolute and normalized
				expected, _ := filepath.Abs(tt.path)
				expected = filepath.Clean(expected)
				if path != expected {
					t.Errorf("Expected path %s, got %s", expected, path)
				}
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Just ensure it doesn't panic
	PrintUsage("1.0.0")
}
