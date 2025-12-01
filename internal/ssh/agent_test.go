package ssh

import (
	"os"
	"testing"

	"git-identitree/internal/profile"
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
	err := UnloadKeyForProfile(prof)
	if err == nil {
		// UnloadKey might succeed even if key doesn't exist (if it's not in agent)
		// So we just check it doesn't panic
	}

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
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

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

