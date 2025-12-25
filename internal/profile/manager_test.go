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

func TestManager_AddProfile_SSHKeyPathWithTilde(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create a test SSH key file in the temp home directory
	sshDir := tmpDir + "/.ssh"
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	keyPath := sshDir + "/id_rsa_test"
	if err := os.WriteFile(keyPath, []byte("test key"), 0600); err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	// Test with tilde path
	profile := Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "~/.ssh/id_rsa_test",
	}

	// This should succeed because ExpandPath will expand ~ to tmpDir
	if err := manager.AddProfile(profile); err != nil {
		t.Fatalf("AddProfile() with tilde path error = %v", err)
	}

	// Verify profile was added with tilde path (not expanded in storage)
	got, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if got.SSHKeyPath != "~/.ssh/id_rsa_test" {
		t.Errorf("Profile SSHKeyPath = %v, want ~/ .ssh/id_rsa_test", got.SSHKeyPath)
	}
}

func TestManager_AddProfile_SSHKeyPathWithTilde_NonExistent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Test with tilde path that doesn't exist
	profile := Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "~/.ssh/nonexistent_key",
	}

	// This should fail because the expanded path doesn't exist
	if err := manager.AddProfile(profile); err == nil {
		t.Error("AddProfile() should fail for non-existent SSH key even with tilde")
	}
}

func TestManager_UpdateProfile_SSHKeyPathWithTilde(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Add initial profile without SSH key
	profile := Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(profile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Create a test SSH key file
	sshDir := tmpDir + "/.ssh"
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	keyPath := sshDir + "/id_rsa_updated"
	if err := os.WriteFile(keyPath, []byte("test key"), 0600); err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	// Update with tilde path
	updated := Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "~/.ssh/id_rsa_updated",
	}

	if err := manager.UpdateProfile("test", updated); err != nil {
		t.Fatalf("UpdateProfile() with tilde path error = %v", err)
	}

	// Verify profile was updated
	got, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if got.SSHKeyPath != "~/.ssh/id_rsa_updated" {
		t.Errorf("Profile SSHKeyPath = %v, want ~/.ssh/id_rsa_updated", got.SSHKeyPath)
	}
}
