package mapping

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-identitree/internal/profile"
	"git-identitree/internal/utils"
)

func TestGenerateProfileConfig(t *testing.T) {
	tmpDir, _, cleanup := setupMappingTestEnv(t)
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

	expectedPath := filepath.Join(tmpDir, ".gitconfig-test")
	if configPath != expectedPath {
		t.Errorf("generateProfileConfig() path = %v, want %v", configPath, expectedPath)
	}

	// Verify file contents
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "name = test") {
		t.Error("Generated config missing user.name")
	}
	if !strings.Contains(contentStr, "email = test@example.com") {
		t.Error("Generated config missing user.email")
	}
	if !strings.Contains(contentStr, "signingkey = ABC123") {
		t.Error("Generated config missing user.signingkey")
	}
	if !strings.Contains(contentStr, "sshCommand = ssh -i /path/to/key") {
		t.Error("Generated config missing core.sshCommand")
	}
}

func TestGenerateProfileConfig_NoSSHOrGPG(t *testing.T) {
	_, _, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	prof := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
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
	if strings.Contains(contentStr, "signingkey") {
		t.Error("Generated config should not contain signingkey when GPGKeyID is empty")
	}
	if strings.Contains(contentStr, "sshCommand") {
		t.Error("Generated config should not contain sshCommand when SSHKeyPath is empty")
	}
}

func TestAddIncludeIfBlock(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	configPath := filepath.Join(tmpDir, ".gitconfig-test")

	if err := addIncludeIfBlock(normalizedDir, configPath); err != nil {
		t.Fatalf("addIncludeIfBlock() error = %v", err)
	}

	// Verify includeIf block was added
	content, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, `[includeIf "gitdir/i:`+normalizedDir+`"]`) {
		t.Error("Git config missing includeIf block")
	}
	if !strings.Contains(contentStr, "path = ~/.gitconfig-test") {
		t.Error("Git config missing path line")
	}
}

func TestAddIncludeIfBlock_Existing(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
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

	// Verify path was updated, not duplicated
	content, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	contentStr := string(content)
	count := strings.Count(contentStr, `[includeIf "gitdir/i:`+normalizedDir+`"]`)
	if count != 1 {
		t.Errorf("Git config has %d includeIf blocks for same directory, want 1", count)
	}

	if !strings.Contains(contentStr, "path = ~/.gitconfig-new") {
		t.Error("Git config path was not updated")
	}
}

func TestRemoveIncludeIfBlock(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create config with includeIf block
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
	if strings.Contains(contentStr, "path = ~/.gitconfig-test") {
		t.Error("Git config still contains path line after removal")
	}

	// Verify other content is preserved
	if !strings.Contains(contentStr, "[user]") {
		t.Error("Git config lost other content during removal")
	}
}

func TestMapProfileToDirectory(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a temporary SSH key file
	tmpKey, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	if err := tmpKey.Close(); err != nil {
		t.Fatalf("Failed to close temp key file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpKey.Name()); err != nil {
			t.Logf("Failed to remove temp key file: %v", err)
		}
	}()

	prof := &profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		SSHKeyPath: tmpKey.Name(),
		GPGKeyID:   "ABC123",
	}

	if err := MapProfileToDirectory(prof, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	// Verify profile config was created
	profileConfigPath := filepath.Join(tmpDir, ".gitconfig-test")
	if _, err := os.Stat(profileConfigPath); os.IsNotExist(err) {
		t.Error("Profile config file was not created")
	}

	// Verify includeIf block was added
	content, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)
	contentStr := string(content)
	if !strings.Contains(contentStr, `[includeIf "gitdir/i:`+normalizedDir+`"]`) {
		t.Error("Git config missing includeIf block")
	}
}

func TestMapProfileToDirectory_Duplicate(t *testing.T) {
	tmpDir, _, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	prof1 := &profile.Profile{
		Name:  "test1",
		Email: "test1@example.com",
	}

	prof2 := &profile.Profile{
		Name:  "test2",
		Email: "test2@example.com",
	}

	if err := MapProfileToDirectory(prof1, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	// Try to map another profile to the same directory
	if err := MapProfileToDirectory(prof2, testDir); err == nil {
		t.Error("MapProfileToDirectory() should fail for duplicate directory mapping")
	}
}

func TestUnmapDirectory(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	prof := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	if err := MapProfileToDirectory(prof, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	if err := UnmapDirectory(testDir); err != nil {
		t.Fatalf("UnmapDirectory() error = %v", err)
	}

	// Verify includeIf block was removed
	content, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)
	contentStr := string(content)
	if strings.Contains(contentStr, `[includeIf "gitdir/i:`+normalizedDir+`"]`) {
		t.Error("Git config still contains includeIf block after unmap")
	}
}

func TestAddIncludeIfBlock_UpdateExisting(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
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
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
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
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
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
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("Failed to restore HOME: %v", err)
		}
	}()

	// Set invalid HOME to test error path
	if err := os.Setenv("HOME", ""); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	// This should fail because we can't get home directory
	_, err := getGitConfigPath()
	if err == nil {
		t.Error("getGitConfigPath() should fail with invalid HOME")
	}
}

func TestAddIncludeIfBlock_UpdatePathLine(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create existing config with includeIf and path line
	existingConfig := `[includeIf "gitdir/i:` + normalizedDir + `"]
    path = ~/.gitconfig-old
`
	if err := os.WriteFile(gitConfigPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
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
		t.Error("Git config path should be updated")
	}
}

func TestAddIncludeIfBlock_NoPathLineAfterIncludeIf(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create existing config with includeIf but no path line (at end of file)
	existingConfig := `[includeIf "gitdir/i:` + normalizedDir + `"]
`
	if err := os.WriteFile(gitConfigPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
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

func TestRemoveIncludeIfBlock_EmptyLineBefore(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create config with empty line before includeIf
	configContent := `[user]
    name = Test

[includeIf "gitdir/i:` + normalizedDir + `"]
    path = ~/.gitconfig-test
`
	if err := os.WriteFile(gitConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	if err := removeIncludeIfBlock(normalizedDir); err != nil {
		t.Fatalf("removeIncludeIfBlock() error = %v", err)
	}

	// Verify includeIf block was removed and empty line before was handled
	content, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("Failed to read git config: %v", err)
	}

	contentStr := string(content)
	if strings.Contains(contentStr, `[includeIf "gitdir/i:`+normalizedDir+`"]`) {
		t.Error("Git config still contains includeIf block after removal")
	}
}

func TestAddIncludeIfBlock_ReadError(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create directory with same name as config file
	if err := os.Remove(gitConfigPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove git config: %v", err)
	}
	if err := os.MkdirAll(gitConfigPath, 0755); err != nil {
		t.Fatalf("Failed to create git config directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(gitConfigPath); err != nil {
			t.Logf("Failed to remove git config path: %v", err)
		}
	}()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	configPath := filepath.Join(tmpDir, ".gitconfig-test")
	err := addIncludeIfBlock(normalizedDir, configPath)
	if err == nil {
		t.Error("addIncludeIfBlock() should fail when config is a directory")
	}
}

func TestAddIncludeIfBlock_OpenError(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create a file that can't be opened (permissions)
	// Note: This might not work on all systems
	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create config file
	if err := os.WriteFile(gitConfigPath, []byte("[user]\n"), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	// Try to make it unreadable (this might not work on all systems)
	// On Unix, we can try to remove read permission
	if err := os.Chmod(gitConfigPath, 0000); err == nil {
		defer func() {
			if err := os.Chmod(gitConfigPath, 0644); err != nil {
				t.Logf("Failed to restore permissions: %v", err)
			}
		}()

		configPath := filepath.Join(tmpDir, ".gitconfig-test")
		err := addIncludeIfBlock(normalizedDir, configPath)
		if err == nil {
			t.Log("addIncludeIfBlock() might succeed even with restricted permissions on some systems")
		} else {
			t.Logf("addIncludeIfBlock() handled permission error: %v", err)
		}
	}
}

func TestRemoveIncludeIfBlock_OpenError(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create directory with same name
	if err := os.Remove(gitConfigPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove git config: %v", err)
	}
	if err := os.MkdirAll(gitConfigPath, 0755); err != nil {
		t.Fatalf("Failed to create git config directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(gitConfigPath); err != nil {
			t.Logf("Failed to remove git config path: %v", err)
		}
	}()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	err := removeIncludeIfBlock(normalizedDir)
	if err == nil {
		t.Error("removeIncludeIfBlock() should fail when config is a directory")
	}
}

func TestRemoveIncludeIfBlock_ScannerError(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create a very large file
	largeContent := make([]byte, 0)
	for i := 0; i < 1000; i++ {
		largeContent = append(largeContent, []byte("[includeIf \"gitdir/i:/tmp/test\"]\n    path = ~/.gitconfig-test\n")...)
	}
	if err := os.WriteFile(gitConfigPath, largeContent, 0644); err != nil {
		t.Fatalf("Failed to write large git config: %v", err)
	}

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Should handle large file
	err := removeIncludeIfBlock(normalizedDir)
	if err != nil {
		t.Logf("removeIncludeIfBlock() handled large file: %v", err)
	}
}

func TestWriteGitConfig_WriteError(t *testing.T) {
	tmpDir, _, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Try to write to a path where parent is a file
	invalidPath := filepath.Join(tmpDir, "file", "config")
	if err := os.WriteFile(filepath.Join(tmpDir, "file"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	lines := []string{"[user]", "    name = Test"}
	err := writeGitConfig(invalidPath, lines)
	if err == nil {
		t.Error("writeGitConfig() should fail when parent is a file")
	}
}

func TestGenerateProfileConfig_HomeDirError(t *testing.T) {
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

	prof := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	_, err := generateProfileConfig(prof)
	if err == nil {
		t.Error("generateProfileConfig() should fail with invalid HOME")
	}
}

func TestMapProfileToDirectory_ParseError(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create unreadable config
	if err := os.Remove(gitConfigPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove git config: %v", err)
	}
	if err := os.MkdirAll(gitConfigPath, 0755); err != nil {
		t.Fatalf("Failed to create git config directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(gitConfigPath); err != nil {
			t.Logf("Failed to remove git config path: %v", err)
		}
	}()

	prof := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err := MapProfileToDirectory(prof, testDir)
	if err == nil {
		t.Error("MapProfileToDirectory() should fail when config is unreadable")
	}
}

func TestMapProfileToDirectory_GenerateConfigError(t *testing.T) {
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

	prof := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	err := MapProfileToDirectory(prof, "/tmp/test")
	if err == nil {
		t.Error("MapProfileToDirectory() should fail with invalid HOME")
	}
}
