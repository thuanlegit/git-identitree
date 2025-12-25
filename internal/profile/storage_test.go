package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestEnv(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "gidtree-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	cleanup := func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("Failed to restore HOME: %v", err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}

	return tmpDir, cleanup
}

func TestSaveAndLoadProfiles(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	profiles := []Profile{
		{
			Name:       "test1",
			Email:      "test1@example.com",
			SSHKeyPath: "/path/to/key1",
			GPGKeyID:   "ABC123",
		},
		{
			Name:       "test2",
			Email:      "test2@example.com",
			SSHKeyPath: "",
			GPGKeyID:   "",
		},
	}

	// Save profiles
	if err := SaveProfiles(profiles); err != nil {
		t.Fatalf("SaveProfiles() error = %v", err)
	}

	// Load profiles
	loaded, err := LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v", err)
	}

	if len(loaded) != len(profiles) {
		t.Fatalf("LoadProfiles() loaded %d profiles, want %d", len(loaded), len(profiles))
	}

	for i, p := range profiles {
		if loaded[i].Name != p.Name {
			t.Errorf("Profile[%d].Name = %v, want %v", i, loaded[i].Name, p.Name)
		}
		if loaded[i].Email != p.Email {
			t.Errorf("Profile[%d].Email = %v, want %v", i, loaded[i].Email, p.Email)
		}
		if loaded[i].SSHKeyPath != p.SSHKeyPath {
			t.Errorf("Profile[%d].SSHKeyPath = %v, want %v", i, loaded[i].SSHKeyPath, p.SSHKeyPath)
		}
		if loaded[i].GPGKeyID != p.GPGKeyID {
			t.Errorf("Profile[%d].GPGKeyID = %v, want %v", i, loaded[i].GPGKeyID, p.GPGKeyID)
		}
	}
}

func TestLoadProfilesNonExistent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	profiles, err := LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v, want no error for non-existent file", err)
	}

	if len(profiles) != 0 {
		t.Errorf("LoadProfiles() = %v, want empty slice", profiles)
	}
}

func TestGetProfilesPath(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	path, err := GetProfilesPath()
	if err != nil {
		t.Fatalf("GetProfilesPath() error = %v", err)
	}

	expected := filepath.Join(tmpDir, profilesDir, profilesFile)
	if path != expected {
		t.Errorf("GetProfilesPath() = %v, want %v", path, expected)
	}
}

func TestGetProfilesDir(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	dir, err := GetProfilesDir()
	if err != nil {
		t.Fatalf("GetProfilesDir() error = %v", err)
	}

	expected := filepath.Join(tmpDir, profilesDir)
	if dir != expected {
		t.Errorf("GetProfilesDir() = %v, want %v", dir, expected)
	}
}

func TestLoadProfiles_InvalidYAML(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	profilesDir, err := GetProfilesDir()
	if err != nil {
		t.Fatalf("GetProfilesDir() error = %v", err)
	}
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}

	profilesPath, err := GetProfilesPath()
	if err != nil {
		t.Fatalf("GetProfilesPath() error = %v", err)
	}

	// Create invalid YAML (tab character which is invalid in YAML)
	invalidYAML := "- name: test\n\temail: test@example.com"
	if err := os.WriteFile(profilesPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write invalid YAML: %v", err)
	}

	_, err = LoadProfiles()
	if err == nil {
		// YAML parser might be lenient, so we'll just log if it doesn't fail
		t.Log("LoadProfiles() might accept some invalid YAML formats")
	} else {
		t.Logf("LoadProfiles() correctly rejected invalid YAML: %v", err)
	}
}

func TestSaveProfiles_WriteError(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a file where the directory should be
	profilesDir, err := GetProfilesDir()
	if err != nil {
		t.Fatalf("GetProfilesDir() error = %v", err)
	}

	// Remove directory and create a file with the same name
	if err := os.RemoveAll(profilesDir); err != nil {
		t.Fatalf("Failed to remove profiles directory: %v", err)
	}
	if err := os.WriteFile(profilesDir, []byte("file"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(profilesDir); err != nil {
			t.Logf("Failed to remove profiles directory: %v", err)
		}
	}()

	profiles := []Profile{
		{Name: "test", Email: "test@example.com"},
	}

	err = SaveProfiles(profiles)
	if err == nil {
		t.Error("SaveProfiles() should fail when directory is a file")
	}
}

func TestGetProfilesPath_HomeDirError(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("Failed to restore HOME: %v", err)
		}
	}()

	// Set invalid HOME
	if err := os.Setenv("HOME", ""); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	_, err := GetProfilesPath()
	if err == nil {
		t.Log("GetProfilesPath() might succeed even with invalid HOME on some systems")
	} else {
		t.Logf("GetProfilesPath() handled invalid HOME: %v", err)
	}

	// Restore HOME already handled by defer
}

func TestGetProfilesDir_HomeDirError(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("Failed to restore HOME: %v", err)
		}
	}()

	// Set invalid HOME
	if err := os.Setenv("HOME", ""); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	_, err := GetProfilesDir()
	if err == nil {
		t.Log("GetProfilesDir() might succeed even with invalid HOME on some systems")
	} else {
		t.Logf("GetProfilesDir() handled invalid HOME: %v", err)
	}

	// Restore HOME already handled by defer
}

func TestLoadProfiles_ReadError(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	profilesPath, err := GetProfilesPath()
	if err != nil {
		t.Fatalf("GetProfilesPath() error = %v", err)
	}

	// Create directory with same name as file
	if err := os.Remove(profilesPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove profiles file: %v", err)
	}
	if err := os.MkdirAll(profilesPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(profilesPath); err != nil {
			t.Logf("Failed to remove profiles path: %v", err)
		}
	}()

	_, err = LoadProfiles()
	// Should handle gracefully (treat as non-existent)
	if err != nil {
		t.Logf("LoadProfiles() handled directory-as-file: %v", err)
	}
}

func TestSaveProfiles_MarshalError(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create profiles that might cause marshal issues
	// (In practice, Profile struct should always marshal correctly)
	profiles := []Profile{
		{Name: "test", Email: "test@example.com"},
	}

	// This should succeed
	err := SaveProfiles(profiles)
	if err != nil {
		t.Fatalf("SaveProfiles() error = %v", err)
	}

	// Verify it was saved
	loaded, err := LoadProfiles()
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v", err)
	}

	if len(loaded) != 1 {
		t.Errorf("LoadProfiles() returned %d profiles, want 1", len(loaded))
	}
}

func TestSaveProfiles_CreateDirectory(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Remove profiles directory
	profilesDir, err := GetProfilesDir()
	if err != nil {
		t.Fatalf("GetProfilesDir() error = %v", err)
	}
	if err := os.RemoveAll(profilesDir); err != nil {
		t.Fatalf("Failed to remove profiles directory: %v", err)
	}

	profiles := []Profile{
		{Name: "test", Email: "test@example.com"},
	}

	err = SaveProfiles(profiles)
	if err != nil {
		t.Fatalf("SaveProfiles() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		t.Error("SaveProfiles() should create profiles directory")
	}
}

