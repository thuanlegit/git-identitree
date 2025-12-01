package profile

import (
	"os"
	"testing"
)

func TestManager_UpdateProfile_NonExistent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	updated := Profile{
		Name:  "nonexistent",
		Email: "test@example.com",
	}

	if err := manager.UpdateProfile("nonexistent", updated); err == nil {
		t.Error("UpdateProfile() should fail for non-existent profile")
	}
}

func TestManager_UpdateProfile_InvalidSSHKey(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile := Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(profile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	updated := Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "/nonexistent/key",
	}

	if err := manager.UpdateProfile("test", updated); err == nil {
		t.Error("UpdateProfile() should fail for non-existent SSH key")
	}
}

func TestManager_DeleteProfile_NoCheck(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile := Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(profile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Delete with nil check function
	if err := manager.DeleteProfile("test", nil); err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
	}

	// Verify deleted
	_, err = manager.GetProfile("test")
	if err == nil {
		t.Error("Profile should have been deleted")
	}
}

func TestManager_DeleteProfile_CheckError(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile := Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(profile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Delete with check function that returns error
	isMapped := func(name string) (bool, error) {
		return false, os.ErrPermission
	}

	if err := manager.DeleteProfile("test", isMapped); err == nil {
		t.Error("DeleteProfile() should fail when check function returns error")
	}
}

func TestManager_AddProfile_EmptyName(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// This would be caught by validation in the form, but test the manager logic
	profile := Profile{
		Name:  "",
		Email: "test@example.com",
	}

	// Manager doesn't validate empty name, but it's a valid test case
	if err := manager.AddProfile(profile); err != nil {
		t.Logf("AddProfile() handled empty name: %v", err)
	}
}

