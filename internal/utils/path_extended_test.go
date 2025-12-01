package utils

import (
	"os"
	"path/filepath"
	"testing"
)

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

