package mapping

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-identitree/internal/profile"
	"git-identitree/internal/utils"
)

func TestAddIncludeIfBlock_UpdateExisting(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(testDir, 0755)
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create existing config with includeIf block
	existingConfig := `[user]
    name = Test

[includeIf "gitdir/i:` + normalizedDir + `"]
    path = ~/.gitconfig-old
`
	if err := os.WriteFile(gitConfigPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to write existing config: %v", err)
	}

	newConfigPath := filepath.Join(tmpDir, ".gitconfig-new")
	if err := addIncludeIfBlock(normalizedDir, newConfigPath); err != nil {
		t.Fatalf("addIncludeIfBlock() error = %v", err)
	}

	// Verify path was updated
	content, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "path = ~/.gitconfig-new") {
		t.Error("Git config path was not updated")
	}
}

func TestAddIncludeIfBlock_NoPathLine(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(testDir, 0755)
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create existing config with includeIf but no path line
	existingConfig := `[includeIf "gitdir/i:` + normalizedDir + `"]
    other = value
`
	if err := os.WriteFile(gitConfigPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to write existing config: %v", err)
	}

	newConfigPath := filepath.Join(tmpDir, ".gitconfig-new")
	if err := addIncludeIfBlock(normalizedDir, newConfigPath); err != nil {
		t.Fatalf("addIncludeIfBlock() error = %v", err)
	}

	// Should append new block
	content, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	contentStr := string(content)
	// Should have the new path line
	if !strings.Contains(contentStr, "path = ~/.gitconfig-new") {
		t.Error("Git config should have new path line")
	}
}

func TestRemoveIncludeIfBlock_WithEmptyLineBefore(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(testDir, 0755)
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create config with empty line before includeIf
	configContent := `[user]
    name = Test

[includeIf "gitdir/i:` + normalizedDir + `"]
    path = ~/.gitconfig-test
`
	if err := os.WriteFile(gitConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	if err := removeIncludeIfBlock(normalizedDir); err != nil {
		t.Fatalf("removeIncludeIfBlock() error = %v", err)
	}

	// Verify includeIf block was removed
	content, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	contentStr := string(content)
	if strings.Contains(contentStr, `[includeIf "gitdir/i:`+normalizedDir+`"]`) {
		t.Error("Git config still contains includeIf block after removal")
	}
}

func TestWriteGitConfig_CreateParentDir(t *testing.T) {
	tmpDir, _, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Write to a nested path
	nestedPath := filepath.Join(tmpDir, "nested", "dir", ".gitconfig")
	lines := []string{"[user]", "    name = Test"}

	if err := writeGitConfig(nestedPath, lines); err != nil {
		t.Fatalf("writeGitConfig() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("writeGitConfig() did not create file in nested directory")
	}
}

func TestGenerateProfileConfig_AllFields(t *testing.T) {
	_, _, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	prof := &profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: "/path/to/key",
		GPGKeyID:   "ABC123",
	}

	configPath, err := generateProfileConfig(prof)
	if err != nil {
		t.Fatalf("generateProfileConfig() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
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
			t.Errorf("Generated config missing: %s", check)
		}
	}
}

func TestMapProfileToDirectory_ErrorPaths(t *testing.T) {
	_, _, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	prof := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	// Test with invalid directory path (should normalize but might fail)
	// Using a relative path that will be normalized
	testDir := "relative/path"
	
	// This should work after normalization
	err := MapProfileToDirectory(prof, testDir)
	if err != nil {
		t.Logf("MapProfileToDirectory() handled relative path: %v", err)
	}
}

func TestUnmapDirectory_NonExistent(t *testing.T) {
	_, _, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Try to unmap a directory that was never mapped
	testDir := "/nonexistent/directory"
	
	// Should not error, just do nothing
	err := UnmapDirectory(testDir)
	if err != nil {
		t.Logf("UnmapDirectory() handled non-existent mapping: %v", err)
	}
}

func TestGetGitConfigPath_Error(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set invalid HOME to test error path
	os.Setenv("HOME", "")

	// This should fail because we can't get home directory
	_, err := getGitConfigPath()
	if err == nil {
		t.Error("getGitConfigPath() should fail with invalid HOME")
	}

	// Restore HOME
	os.Setenv("HOME", originalHome)
}

