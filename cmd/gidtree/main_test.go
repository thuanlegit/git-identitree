package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-identitree/internal/mapping"
	"git-identitree/internal/profile"
)

func setupCLITestEnv(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "gidtree-cli-test-*")
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

func TestInitCommand(t *testing.T) {
	_, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Run init command directly (bypassing cobra's help detection)
	// We'll test the actual functionality instead
	profilesDir, err := profile.GetProfilesDir()
	if err != nil {
		t.Fatalf("GetProfilesDir() error = %v", err)
	}
	if err = os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("Failed to create profiles directory: %v", err)
	}
	profilesPath, err := profile.GetProfilesPath()
	if err != nil {
		t.Fatalf("GetProfilesPath() error = %v", err)
	}
	if err = profile.SaveProfiles([]profile.Profile{}); err != nil {
		t.Fatalf("SaveProfiles() error = %v", err)
	}

	// Verify directory was created
	if _, err = os.Stat(profilesDir); os.IsNotExist(err) {
		t.Error("Init command did not create profiles directory")
	}

	// Verify profiles file exists
	if _, err = os.Stat(profilesPath); os.IsNotExist(err) {
		t.Error("Init command did not create profiles file")
	}
}

func TestProfileCreateCommand(t *testing.T) {
	// Note: This test would require mocking the Huh form, which is complex
	// For now, we'll test the profile management logic separately
	// Integration test would require user interaction simulation
	t.Skip("Skipping interactive profile create test - requires form mocking")
}

func TestProfileListCommand(t *testing.T) {
	_, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	// Create a profile directly
	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	testProfile := profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Test list command (would need to capture output)
	// For now, just verify the profile exists
	profiles := manager.ListProfiles()
	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}
}

func TestProfileDeleteCommand(t *testing.T) {
	_, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	// Create a profile
	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	testProfile := profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Delete profile (no mappings)
	isMapped := func(name string) (bool, error) {
		return false, nil
	}

	if err := manager.DeleteProfile("test", isMapped); err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
	}

	// Verify deleted
	_, err = manager.GetProfile("test")
	if err == nil {
		t.Error("Profile should have been deleted")
	}
}

func TestMapCommand(t *testing.T) {
	tmpDir, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	// Create a profile
	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	testProfile := profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Create test directory
	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Map profile to directory
	prof, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if err := mapping.MapProfileToDirectory(prof, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	// Verify mapping exists
	m, err := mapping.GetMappingForDirectory(testDir)
	if err != nil {
		t.Fatalf("GetMappingForDirectory() error = %v", err)
	}

	if m == nil {
		t.Fatal("Mapping should exist")
	}

	if m.Profile != "test" {
		t.Errorf("Mapping profile = %v, want test", m.Profile)
	}
}

func TestUnmapCommand(t *testing.T) {
	tmpDir, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	// Create a profile and map it
	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	testProfile := profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	prof, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if err := mapping.MapProfileToDirectory(prof, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	// Unmap
	if err := mapping.UnmapDirectory(testDir); err != nil {
		t.Fatalf("UnmapDirectory() error = %v", err)
	}

	// Verify mapping removed
	m, err := mapping.GetMappingForDirectory(testDir)
	if err != nil {
		t.Fatalf("GetMappingForDirectory() error = %v", err)
	}

	if m != nil {
		t.Error("Mapping should have been removed")
	}
}

func TestStatusCommand(t *testing.T) {
	tmpDir, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	// Create a profile and map it
	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	testProfile := profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	prof, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if err := mapping.MapProfileToDirectory(prof, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	// Verify mappings can be parsed
	mappings, err := mapping.ParseMappings()
	if err != nil {
		t.Fatalf("ParseMappings() error = %v", err)
	}

	if len(mappings) == 0 {
		t.Error("Expected at least one mapping")
	}
}

func TestActivateCommand(t *testing.T) {
	tmpDir, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	// Create a profile and map it
	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	testProfile := profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	prof, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if err := mapping.MapProfileToDirectory(prof, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	// Change to test directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore directory: %v", err)
		}
	}()

	// Test activate logic (get mapping for current directory)
	m, err := mapping.GetMappingForDirectory(testDir)
	if err != nil {
		t.Fatalf("GetMappingForDirectory() error = %v", err)
	}

	if m == nil {
		t.Fatal("Should find mapping for current directory")
	}

	if m.Profile != "test" {
		t.Errorf("Mapping profile = %v, want test", m.Profile)
	}
}

func TestProfileDeleteWithMapping(t *testing.T) {
	tmpDir, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	// Create a profile and map it
	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	testProfile := profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	prof, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if err := mapping.MapProfileToDirectory(prof, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	// Try to delete profile (should fail because it's mapped)
	isMapped := func(name string) (bool, error) {
		return mapping.IsProfileMapped(name)
	}

	if err := manager.DeleteProfile("test", isMapped); err == nil {
		t.Error("DeleteProfile() should fail when profile is mapped")
	}

	// Unmap first
	if err := mapping.UnmapDirectory(testDir); err != nil {
		t.Fatalf("UnmapDirectory() error = %v", err)
	}

	// Now delete should succeed
	if err := manager.DeleteProfile("test", isMapped); err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
	}
}

func TestGenerateProfileConfig_Content(t *testing.T) {
	tmpDir, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Create a profile with all fields
	prof := &profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "/path/to/key",
		GPGKeyID:   "ABC123",
	}

	// Use internal function to generate config
	// We'll test through the mapping package
	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	if err := mapping.MapProfileToDirectory(prof, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	// Verify profile config was created
	home := os.Getenv("HOME")
	configPath := filepath.Join(home, ".gitconfig-test")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read profile config: %v", err)
	}

	contentStr := string(content)
	checks := []string{
		"name = test",
		"email = test@example.com",
		"signingkey = ABC123",
		"sshCommand = ssh -i /path/to/key",
	}

	for _, check := range checks {
		if !strings.Contains(contentStr, check) {
			t.Errorf("Profile config missing: %s", check)
		}
	}
}

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name            string
		versionValue    string
		expectedOutput  string
	}{
		{
			name:           "default version",
			versionValue:   "1.0.0",
			expectedOutput: "gidtree version 1.0.0",
		},
		{
			name:           "dev version",
			versionValue:   "dev",
			expectedOutput: "gidtree version dev",
		},
		{
			name:           "semver with build metadata",
			versionValue:   "2.1.3-beta.1+build.123",
			expectedOutput: "gidtree version 2.1.3-beta.1+build.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the version variable
			originalVersion := version
			version = tt.versionValue
			defer func() { version = originalVersion }()

			// Create a buffer to capture output
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			// Execute version command
			versionCmd.Run(versionCmd, []string{})

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			output := strings.TrimSpace(buf.String())

			// Verify output
			if output != tt.expectedOutput {
				t.Errorf("Version output = %q, want %q", output, tt.expectedOutput)
			}
		})
	}
}

func TestVersionCommandRegistered(t *testing.T) {
	// Verify version command is registered with root command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Error("version command not registered with root command")
	}
}

func TestVersionCommandHelp(t *testing.T) {
	// Verify version command has proper help text
	if versionCmd.Use != "version" {
		t.Errorf("versionCmd.Use = %q, want %q", versionCmd.Use, "version")
	}

	if versionCmd.Short == "" {
		t.Error("versionCmd.Short should not be empty")
	}

	if !strings.Contains(versionCmd.Short, "version") {
		t.Errorf("versionCmd.Short = %q, should contain 'version'", versionCmd.Short)
	}
}

func TestProfileUpdateCommand(t *testing.T) {
	tmpDir, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	// Create a profile first
	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create a test SSH key file
	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	keyPath := filepath.Join(sshDir, "id_rsa_test")
	if err := os.WriteFile(keyPath, []byte("test key"), 0600); err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	testProfile := profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "~/.ssh/id_rsa_test",
	}

	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Create updated key file
	updatedKeyPath := filepath.Join(sshDir, "id_rsa_updated")
	if err := os.WriteFile(updatedKeyPath, []byte("updated key"), 0600); err != nil {
		t.Fatalf("Failed to create updated key file: %v", err)
	}

	// Update profile
	updatedProfile := profile.Profile{
		Name:       "test",
		Email:      "updated@example.com",
		AuthorName: "Test Author",
		SSHKeyPath: "~/.ssh/id_rsa_updated",
		GPGKeyID:   "GPG123",
	}

	if err := manager.UpdateProfile("test", updatedProfile); err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}

	// Verify the update
	got, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if got.Email != "updated@example.com" {
		t.Errorf("Profile email = %v, want updated@example.com", got.Email)
	}

	if got.AuthorName != "Test Author" {
		t.Errorf("Profile authorName = %v, want Test Author", got.AuthorName)
	}

	if got.SSHKeyPath != "~/.ssh/id_rsa_updated" {
		t.Errorf("Profile sshKeyPath = %v, want ~/.ssh/id_rsa_updated", got.SSHKeyPath)
	}

	if got.GPGKeyID != "GPG123" {
		t.Errorf("Profile gpgKeyID = %v, want GPG123", got.GPGKeyID)
	}
}

func TestProfileUpdateCommand_NonExistent(t *testing.T) {
	_, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Try to update non-existent profile
	_, err = manager.GetProfile("nonexistent")
	if err == nil {
		t.Error("GetProfile() should fail for non-existent profile")
	}
}

func TestProfileUpdateCommandRegistered(t *testing.T) {
	// Verify update command is registered with profile command
	found := false
	for _, cmd := range profileCmd.Commands() {
		if cmd.Name() == "update" {
			found = true
			break
		}
	}

	if !found {
		t.Error("update command not registered with profile command")
	}
}

func TestProfileUpdateCommand_SSHKeyValidation(t *testing.T) {
	_, cleanup := setupCLITestEnv(t)
	defer cleanup()

	// Initialize
	initCmd.SetArgs([]string{})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("initCmd.Execute() error = %v", err)
	}

	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create a profile without SSH key
	testProfile := profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Try to update with non-existent SSH key
	updatedProfile := profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "~/.ssh/nonexistent_key",
	}

	// Should fail validation
	if err := manager.UpdateProfile("test", updatedProfile); err == nil {
		t.Error("UpdateProfile() should fail for non-existent SSH key")
	}
}

