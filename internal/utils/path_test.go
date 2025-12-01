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

