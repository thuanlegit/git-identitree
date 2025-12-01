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
	os.MkdirAll(testDir, 0755)
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
	os.MkdirAll(testDir, 0755)
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
	os.MkdirAll(testDir, 0755)

	// Create a temporary SSH key file
	tmpKey, err := os.CreateTemp("", "test-key-*")
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	tmpKey.Close()
	defer os.Remove(tmpKey.Name())

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
	os.MkdirAll(testDir, 0755)

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
	os.MkdirAll(testDir, 0755)

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
