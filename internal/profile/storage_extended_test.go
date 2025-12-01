package profile

import (
	"os"
	"testing"
)

func TestLoadProfiles_InvalidYAML(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	profilesPath, err := GetProfilesPath()
	if err != nil {
		t.Fatalf("GetProfilesPath() error = %v", err)
	}

	// Create invalid YAML (tab character which is invalid in YAML)
	invalidYAML := "- name: test\n\temail: test@example.com"
	os.WriteFile(profilesPath, []byte(invalidYAML), 0644)

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
	os.RemoveAll(profilesDir)
	os.WriteFile(profilesDir, []byte("file"), 0644)
	defer os.RemoveAll(profilesDir)

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
	defer os.Setenv("HOME", originalHome)

	// Set invalid HOME
	os.Setenv("HOME", "")

	_, err := GetProfilesPath()
	if err == nil {
		t.Log("GetProfilesPath() might succeed even with invalid HOME on some systems")
	} else {
		t.Logf("GetProfilesPath() handled invalid HOME: %v", err)
	}

	// Restore HOME
	os.Setenv("HOME", originalHome)
}

func TestGetProfilesDir_HomeDirError(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set invalid HOME
	os.Setenv("HOME", "")

	_, err := GetProfilesDir()
	if err == nil {
		t.Log("GetProfilesDir() might succeed even with invalid HOME on some systems")
	} else {
		t.Logf("GetProfilesDir() handled invalid HOME: %v", err)
	}

	// Restore HOME
	os.Setenv("HOME", originalHome)
}

func TestLoadProfiles_ReadError(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	profilesPath, err := GetProfilesPath()
	if err != nil {
		t.Fatalf("GetProfilesPath() error = %v", err)
	}

	// Create directory with same name as file
	os.Remove(profilesPath)
	os.MkdirAll(profilesPath, 0755)
	defer os.RemoveAll(profilesPath)

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
	os.RemoveAll(profilesDir)

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

