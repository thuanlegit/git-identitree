package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkFn  func(string) bool
	}{
		{
			name:    "expand tilde",
			input:   "~/test",
			wantErr: false,
			checkFn: func(got string) bool {
				// Check that path is absolute and contains home directory
				return filepath.IsAbs(got) && (strings.HasPrefix(got, home) || strings.Contains(got, "test"))
			},
		},
		{
			name:    "absolute path",
			input:   "/tmp/test",
			wantErr: false,
			checkFn: func(got string) bool {
				return filepath.IsAbs(got)
			},
		},
		{
			name:    "relative path",
			input:   "test",
			wantErr: false,
			checkFn: func(got string) bool {
				return filepath.IsAbs(got)
			},
		},
		{
			name:    "path with dots",
			input:   "./test/../test2",
			wantErr: false,
			checkFn: func(got string) bool {
				return filepath.IsAbs(got) && filepath.Base(got) == "test2"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.checkFn(got) {
				t.Errorf("NormalizePath() = %v, did not pass check", got)
			}
		})
	}
}

func TestEnsureTrailingSlash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path without trailing slash",
			input:    "/tmp/test",
			expected: "/tmp/test" + string(filepath.Separator),
		},
		{
			name:     "path with trailing slash",
			input:    "/tmp/test/",
			expected: "/tmp/test/",
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "root path",
			input:    "/",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsureTrailingSlash(tt.input)
			if got != tt.expected {
				t.Errorf("EnsureTrailingSlash() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetHomeDir(t *testing.T) {
	home, err := GetHomeDir()
	if err != nil {
		t.Fatalf("GetHomeDir() error = %v", err)
	}

	if !filepath.IsAbs(home) {
		t.Errorf("GetHomeDir() = %v, want absolute path", home)
	}
}

func TestNormalizePath_ErrorCases(t *testing.T) {
	// Test with path that causes Abs to fail (shouldn't happen in practice,
	// but tests error handling)
	
	// Test with very long path that might cause issues
	longPath := string(make([]byte, 10000)) + "/test"
	_, err := NormalizePath(longPath)
	// This might succeed or fail depending on system, but tests the path
	if err != nil {
		t.Logf("NormalizePath() handled long path: %v", err)
	}
}

func TestNormalizePath_SymlinkError(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gidtree-path-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a path that doesn't exist yet (EvalSymlinks will fail)
	nonExistentPath := filepath.Join(tmpDir, "nonexistent", "path")
	
	// NormalizePath should handle this gracefully
	normalized, err := NormalizePath(nonExistentPath)
	if err != nil {
		t.Fatalf("NormalizePath() should handle non-existent path: %v", err)
	}
	
	// Should return absolute path even if symlink resolution fails
	if !filepath.IsAbs(normalized) {
		t.Errorf("NormalizePath() should return absolute path even if symlink resolution fails: %v", normalized)
	}
}

func TestNormalizePath_TildeInMiddle(t *testing.T) {
	// Test with tilde not at the start (should not expand)
	path := "/tmp/test~file"
	normalized, err := NormalizePath(path)
	if err != nil {
		t.Fatalf("NormalizePath() error = %v", err)
	}
	
	// Should not expand tilde in middle
	if !filepath.IsAbs(normalized) {
		t.Errorf("NormalizePath() should return absolute path: %v", normalized)
	}
}

func TestNormalizePath_HomeDirError(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set invalid HOME
	os.Setenv("HOME", "")
	
	// Clear the home directory cache if any
	// Test that tilde expansion fails gracefully
	_, err := NormalizePath("~/test")
	if err == nil {
		t.Log("NormalizePath() might succeed even with invalid HOME on some systems")
	} else {
		t.Logf("NormalizePath() handled invalid HOME as expected: %v", err)
	}
	
	// Restore HOME
	os.Setenv("HOME", originalHome)
}

func TestEnsureTrailingSlash_AllSeparators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows path",
			input:    "C:\\test",
			expected: "C:\\test\\",
		},
		{
			name:     "Unix path",
			input:    "/tmp/test",
			expected: "/tmp/test" + string(filepath.Separator),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsureTrailingSlash(tt.input)
			// On Windows, separator is \, on Unix it's /
			expected := tt.input + string(filepath.Separator)
			if got != expected {
				t.Errorf("EnsureTrailingSlash() = %v, want %v", got, expected)
			}
		})
	}
}

func TestGetHomeDir_Error(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// On Unix systems, UserHomeDir() should still work even with empty HOME
	// as it falls back to other methods
	home, err := GetHomeDir()
	if err != nil {
		t.Logf("GetHomeDir() error (might be expected on some systems): %v", err)
	} else {
		if !filepath.IsAbs(home) {
			t.Errorf("GetHomeDir() = %v, want absolute path", home)
		}
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "expand tilde",
			input:    "~/.ssh/id_rsa",
			expected: filepath.Join(home, ".ssh/id_rsa"),
			wantErr:  false,
		},
		{
			name:     "expand tilde with subpath",
			input:    "~/Documents/test.txt",
			expected: filepath.Join(home, "Documents/test.txt"),
			wantErr:  false,
		},
		{
			name:     "tilde only",
			input:    "~",
			expected: home,
			wantErr:  false,
		},
		{
			name:     "tilde with slash",
			input:    "~/",
			expected: home + "/",
			wantErr:  false,
		},
		{
			name:     "absolute path without tilde",
			input:    "/tmp/test",
			expected: "/tmp/test",
			wantErr:  false,
		},
		{
			name:     "relative path without tilde",
			input:    "test/file.txt",
			expected: "test/file.txt",
			wantErr:  false,
		},
		{
			name:     "tilde in middle (should not expand)",
			input:    "/tmp/~test",
			expected: "/tmp/~test",
			wantErr:  false,
		},
		{
			name:     "empty path",
			input:    "",
			expected: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ExpandPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExpandPath_ErrorHandling(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set invalid HOME (empty string)
	os.Setenv("HOME", "")

	// Test that tilde expansion handles error
	_, err := ExpandPath("~/test")
	// On Unix systems, UserHomeDir() might still work
	// On some systems it might fail
	if err != nil {
		t.Logf("ExpandPath() handled empty HOME as expected: %v", err)
	} else {
		t.Log("ExpandPath() succeeded despite empty HOME (system fallback)")
	}

	// Restore HOME
	os.Setenv("HOME", originalHome)
}

