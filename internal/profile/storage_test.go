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
	os.Setenv("HOME", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tmpDir)
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

