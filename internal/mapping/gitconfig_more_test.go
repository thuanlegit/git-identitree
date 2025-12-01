package mapping

import (
	"os"
	"path/filepath"
	"testing"

	"git-identitree/internal/profile"
	"git-identitree/internal/utils"
)

func TestAddIncludeIfBlock_ReadError(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create directory with same name as config file
	os.Remove(gitConfigPath)
	os.MkdirAll(gitConfigPath, 0755)
	defer os.RemoveAll(gitConfigPath)

	testDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(testDir, 0755)
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
	os.MkdirAll(testDir, 0755)
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create config file
	os.WriteFile(gitConfigPath, []byte("[user]\n"), 0644)

	// Try to make it unreadable (this might not work on all systems)
	// On Unix, we can try to remove read permission
	if err := os.Chmod(gitConfigPath, 0000); err == nil {
		defer os.Chmod(gitConfigPath, 0644)
		
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
	os.Remove(gitConfigPath)
	os.MkdirAll(gitConfigPath, 0755)
	defer os.RemoveAll(gitConfigPath)

	testDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(testDir, 0755)
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
	os.WriteFile(gitConfigPath, largeContent, 0644)

	testDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(testDir, 0755)
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
	os.WriteFile(filepath.Join(tmpDir, "file"), []byte("content"), 0644)

	lines := []string{"[user]", "    name = Test"}
	err := writeGitConfig(invalidPath, lines)
	if err == nil {
		t.Error("writeGitConfig() should fail when parent is a file")
	}
}

func TestGenerateProfileConfig_HomeDirError(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set invalid HOME
	os.Setenv("HOME", "")

	prof := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	_, err := generateProfileConfig(prof)
	if err == nil {
		t.Error("generateProfileConfig() should fail with invalid HOME")
	}

	// Restore HOME
	os.Setenv("HOME", originalHome)
}

func TestMapProfileToDirectory_ParseError(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create unreadable config
	os.Remove(gitConfigPath)
	os.MkdirAll(gitConfigPath, 0755)
	defer os.RemoveAll(gitConfigPath)

	prof := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	testDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(testDir, 0755)

	err := MapProfileToDirectory(prof, testDir)
	if err == nil {
		t.Error("MapProfileToDirectory() should fail when config is unreadable")
	}
}

func TestMapProfileToDirectory_GenerateConfigError(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set invalid HOME
	os.Setenv("HOME", "")

	prof := &profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}

	err := MapProfileToDirectory(prof, "/tmp/test")
	if err == nil {
		t.Error("MapProfileToDirectory() should fail with invalid HOME")
	}

	// Restore HOME
	os.Setenv("HOME", originalHome)
}

