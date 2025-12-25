package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath converts a path to an absolute, canonical path.
// It resolves ~ to the user's home directory and ensures the path is absolute.
func NormalizePath(path string) (string, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = strings.Replace(path, "~", home, 1)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// Clean the path (remove . and .., resolve symlinks)
	cleanPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If symlink resolution fails, use the absolute path
		// This can happen if the path doesn't exist yet
		cleanPath = absPath
	}

	return cleanPath, nil
}

// EnsureTrailingSlash ensures a directory path ends with a trailing slash.
func EnsureTrailingSlash(path string) string {
	if path == "" {
		return path
	}
	if !strings.HasSuffix(path, string(filepath.Separator)) {
		return path + string(filepath.Separator)
	}
	return path
}

// GetHomeDir returns the user's home directory.
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// ExpandPath expands ~ in a path to the user's home directory.
// Unlike NormalizePath, this does not resolve symlinks or make the path absolute.
func ExpandPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		// Replace only the first ~ with home directory
		path = strings.Replace(path, "~", home, 1)
	}

	return path, nil
}

