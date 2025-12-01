package profile

import (
	"os"
	"testing"
)

func TestNewManager_LoadError(t *testing.T) {
	// This tests the error path in NewManager when LoadProfiles fails
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	profilesPath, err := GetProfilesPath()
	if err != nil {
		t.Fatalf("GetProfilesPath() error = %v", err)
	}

	// Create a file that causes read error (directory)
	os.Remove(profilesPath)
	os.MkdirAll(profilesPath, 0755)
	defer os.RemoveAll(profilesPath)

	_, err = NewManager()
	if err == nil {
		t.Error("NewManager() should fail when profiles file is unreadable")
	}
}

