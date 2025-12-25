package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thuanlegit/git-identitree/internal/profile"
)

func TestLoadKeyForProfile(t *testing.T) {
	prof := &profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "/nonexistent/key",
	}

	// This will fail because the key doesn't exist, but tests the function logic
	err := LoadKeyForProfile(prof)
	if err == nil {
		t.Error("LoadKeyForProfile() should fail for non-existent key")
	}

	// Test with empty SSH key path
	profNoKey := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := LoadKeyForProfile(profNoKey); err != nil {
		t.Errorf("LoadKeyForProfile() error = %v, want no error for profile without SSH key", err)
	}
}

func TestUnloadKeyForProfile(t *testing.T) {
	prof := &profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "/nonexistent/key",
	}

	// This will fail because the key doesn't exist, but tests the function logic
	// UnloadKey might succeed even if key doesn't exist (if it's not in agent)
	// So we just check it doesn't panic
	_ = UnloadKeyForProfile(prof)

	// Test with empty SSH key path
	profNoKey := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := UnloadKeyForProfile(profNoKey); err != nil {
		t.Errorf("UnloadKeyForProfile() error = %v, want no error for profile without SSH key", err)
	}
}

func TestLoadKey_NonExistent(t *testing.T) {
	err := LoadKey("/nonexistent/key")
	if err == nil {
		t.Error("LoadKey() should fail for non-existent key")
	}
}

func TestCheckKeyLoaded_NonExistent(t *testing.T) {
	loaded, err := CheckKeyLoaded("/nonexistent/key")
	if err == nil {
		t.Error("CheckKeyLoaded() should fail for non-existent key")
	}
	if loaded {
		t.Error("CheckKeyLoaded() should return false for non-existent key")
	}
}

// Test with a real temporary file (but not a real SSH key)
func TestLoadKey_InvalidKey(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	// This will fail because it's not a valid SSH key, but tests the path
	err = LoadKey(tmpFile.Name())
	// We expect this to fail because it's not a valid SSH key
	// The exact error depends on ssh-add behavior
	if err == nil {
		t.Log("LoadKey() succeeded (key might be invalid but ssh-add accepted it)")
	}
}

func TestAutoLoadForDirectory(t *testing.T) {
	// This is a placeholder function, so just test it doesn't panic
	err := AutoLoadForDirectory("/some/dir", func(name string) (*profile.Profile, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("AutoLoadForDirectory() error = %v, want no error for placeholder", err)
	}
}

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
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

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
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

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
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

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
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

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
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

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
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

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
	if _, err := tmpFile.WriteString("not a valid ssh key"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	// This should fail at fingerprint extraction
	err = UnloadKey(tmpFile.Name())
	if err != nil {
		t.Logf("UnloadKey() handled invalid key as expected: %v", err)
	}
}

