package mapping

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-identitree/internal/utils"
)

func TestAddIncludeIfBlock_UpdatePathLine(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(testDir, 0755)
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create existing config with includeIf and path line
	existingConfig := `[includeIf "gitdir/i:` + normalizedDir + `"]
    path = ~/.gitconfig-old
`
	os.WriteFile(gitConfigPath, []byte(existingConfig), 0644)

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
	os.MkdirAll(testDir, 0755)
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create existing config with includeIf but no path line (at end of file)
	existingConfig := `[includeIf "gitdir/i:` + normalizedDir + `"]
`
	os.WriteFile(gitConfigPath, []byte(existingConfig), 0644)

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
	os.MkdirAll(testDir, 0755)
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Create config with empty line before includeIf
	configContent := `[user]
    name = Test

[includeIf "gitdir/i:` + normalizedDir + `"]
    path = ~/.gitconfig-test
`
	os.WriteFile(gitConfigPath, []byte(configContent), 0644)

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
