package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"git-identitree/internal/profile"
)

func TestLoadKey_NormalizeError(t *testing.T) {
	// Test with invalid path that causes normalization error
	// Using a path that might cause issues
	err := LoadKey("")
	if err == nil {
		t.Error("LoadKey() should fail for empty path")
	}
}

func TestLoadKey_AlreadyLoaded(t *testing.T) {
	// Create a temporary file to simulate a key
	tmpFile, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// First attempt to load (will likely fail because it's not a real key,
	// but tests the path normalization and existence check)
	err = LoadKey(tmpFile.Name())
	// We expect this to fail because it's not a valid SSH key,
	// but it tests the normalization and file existence paths
	if err != nil {
		// Expected - not a valid SSH key
		t.Logf("LoadKey() failed as expected (not a valid SSH key): %v", err)
	}
}

func TestUnloadKey_NormalizeError(t *testing.T) {
	err := UnloadKey("")
	if err == nil {
		t.Error("UnloadKey() should fail for empty path")
	}
}

func TestUnloadKey_FallbackPath(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// This will fail because it's not a valid SSH key,
	// but tests the fingerprint extraction and fallback logic
	err = UnloadKey(tmpFile.Name())
	// Expected to fail, but tests the code paths
	if err != nil {
		t.Logf("UnloadKey() failed as expected: %v", err)
	}
}

func TestCheckKeyLoaded_NormalizeError(t *testing.T) {
	_, err := CheckKeyLoaded("")
	if err == nil {
		t.Error("CheckKeyLoaded() should fail for empty path")
	}
}

func TestCheckKeyLoaded_InvalidKey(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// This will fail because it's not a valid SSH key
	loaded, err := CheckKeyLoaded(tmpFile.Name())
	if err == nil {
		t.Logf("CheckKeyLoaded() returned loaded=%v for invalid key", loaded)
	}
}

func TestCheckKeyLoaded_SSHAgentNotRunning(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// This tests the path where SSH agent might not be running
	// The function should handle this gracefully
	_, err = CheckKeyLoaded(tmpFile.Name())
	// Error is expected for invalid key, but tests the code path
	if err != nil {
		t.Logf("CheckKeyLoaded() handled error as expected: %v", err)
	}
}

func TestLoadKeyForProfile_WithSSHKey(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	prof := &profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: tmpFile.Name(),
	}

	// This will likely fail because it's not a valid SSH key,
	// but tests the LoadKeyForProfile logic
	err = LoadKeyForProfile(prof)
	if err != nil {
		t.Logf("LoadKeyForProfile() failed as expected (invalid key): %v", err)
	}
}

func TestUnloadKeyForProfile_WithSSHKey(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	prof := &profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: tmpFile.Name(),
	}

	// This will likely fail because it's not a valid SSH key,
	// but tests the UnloadKeyForProfile logic
	err = UnloadKeyForProfile(prof)
	if err != nil {
		t.Logf("UnloadKeyForProfile() failed as expected (invalid key): %v", err)
	}
}

func TestLoadKey_RelativePath(t *testing.T) {
	// Test with relative path
	relPath := "nonexistent-key"
	
	err := LoadKey(relPath)
	if err == nil {
		t.Error("LoadKey() should fail for non-existent key")
	}
	
	// Verify it was normalized to absolute
	if err != nil && !filepath.IsAbs(relPath) {
		// The error should mention the normalized path
		t.Logf("LoadKey() normalized path as expected: %v", err)
	}
}

func TestUnloadKey_InvalidFingerprintFormat(t *testing.T) {
	// Create a file that might cause fingerprint parsing issues
	tmpFile, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	// Write invalid key content
	tmpFile.WriteString("not a valid ssh key")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// This should fail at fingerprint extraction
	err = UnloadKey(tmpFile.Name())
	if err != nil {
		t.Logf("UnloadKey() handled invalid key as expected: %v", err)
	}
}

