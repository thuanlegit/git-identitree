package profile

import (
	"os"
	"testing"
)

func TestManager_AddProfile(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create a temporary SSH key file for testing
	tmpKey, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	tmpKey.Close()
	defer os.Remove(tmpKey.Name())

	profile := Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: tmpKey.Name(),
		GPGKeyID:   "ABC123",
	}

	if err := manager.AddProfile(profile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Try to add duplicate
	if err := manager.AddProfile(profile); err == nil {
		t.Error("AddProfile() should fail for duplicate profile")
	}
}

func TestManager_AddProfile_InvalidSSHKey(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profile := Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "/nonexistent/key",
	}

	if err := manager.AddProfile(profile); err == nil {
		t.Error("AddProfile() should fail for non-existent SSH key")
	}
}

func TestManager_GetProfile(t *testing.T) {
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

	got, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if got.Name != profile.Name {
		t.Errorf("GetProfile().Name = %v, want %v", got.Name, profile.Name)
	}

	// Test non-existent profile
	_, err = manager.GetProfile("nonexistent")
	if err == nil {
		t.Error("GetProfile() should fail for non-existent profile")
	}
}

func TestManager_ListProfiles(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	profiles := []Profile{
		{Name: "test1", Email: "test1@example.com"},
		{Name: "test2", Email: "test2@example.com"},
	}

	for _, p := range profiles {
		if err := manager.AddProfile(p); err != nil {
			t.Fatalf("AddProfile() error = %v", err)
		}
	}

	listed := manager.ListProfiles()
	if len(listed) != len(profiles) {
		t.Errorf("ListProfiles() returned %d profiles, want %d", len(listed), len(profiles))
	}
}

func TestManager_UpdateProfile(t *testing.T) {
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
		Name:  "test",
		Email: "updated@example.com",
	}

	if err := manager.UpdateProfile("test", updated); err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}

	got, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if got.Email != "updated@example.com" {
		t.Errorf("UpdateProfile() email = %v, want updated@example.com", got.Email)
	}
}

func TestManager_DeleteProfile(t *testing.T) {
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

	// Delete with no mappings
	isMapped := func(name string) (bool, error) {
		return false, nil
	}

	if err := manager.DeleteProfile("test", isMapped); err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
	}

	// Verify deleted
	_, err = manager.GetProfile("test")
	if err == nil {
		t.Error("DeleteProfile() should have removed the profile")
	}
}

func TestManager_DeleteProfile_Mapped(t *testing.T) {
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

	// Try to delete with mappings
	isMapped := func(name string) (bool, error) {
		return true, nil
	}

	if err := manager.DeleteProfile("test", isMapped); err == nil {
		t.Error("DeleteProfile() should fail when profile is mapped")
	}
}

func TestManager_DeleteProfile_NonExistent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	isMapped := func(name string) (bool, error) {
		return false, nil
	}

	if err := manager.DeleteProfile("nonexistent", isMapped); err == nil {
		t.Error("DeleteProfile() should fail for non-existent profile")
	}
}

