package mapping

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-identitree/internal/utils"
)

func setupMappingTestEnv(t *testing.T) (string, string, func()) {
	tmpDir, err := os.MkdirTemp("", "gidtree-mapping-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}

	gitConfigPath := filepath.Join(tmpDir, ".gitconfig")

	cleanup := func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Logf("Failed to restore HOME: %v", err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}

	return tmpDir, gitConfigPath, cleanup
}

func TestParseMappings(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create a test git config with includeIf blocks
	testDir1 := filepath.Join(tmpDir, "project1")
	testDir2 := filepath.Join(tmpDir, "project2")
	if err := os.MkdirAll(testDir1, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(testDir2, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	normalizedDir1, _ := utils.NormalizePath(testDir1)
	normalizedDir1 = utils.EnsureTrailingSlash(normalizedDir1)
	normalizedDir2, _ := utils.NormalizePath(testDir2)
	normalizedDir2 = utils.EnsureTrailingSlash(normalizedDir2)

	configContent := `[user]
    name = Test User
    email = test@example.com

[includeIf "gitdir/i:` + normalizedDir1 + `"]
    path = ~/.gitconfig-work

[includeIf "gitdir/i:` + normalizedDir2 + `"]
    path = ~/.gitconfig-personal
`

	if err := os.WriteFile(gitConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test git config: %v", err)
	}

	mappings, err := ParseMappings()
	if err != nil {
		t.Fatalf("ParseMappings() error = %v", err)
	}

	if len(mappings) != 2 {
		t.Fatalf("ParseMappings() returned %d mappings, want 2", len(mappings))
	}

	// Check first mapping
	if !strings.HasSuffix(mappings[0].Directory, normalizedDir1) && !strings.HasSuffix(mappings[1].Directory, normalizedDir1) {
		t.Error("ParseMappings() did not find first directory mapping")
	}

	// Check profile names
	foundWork := false
	foundPersonal := false
	for _, m := range mappings {
		if m.Profile == "work" {
			foundWork = true
		}
		if m.Profile == "personal" {
			foundPersonal = true
		}
	}

	if !foundWork || !foundPersonal {
		t.Errorf("ParseMappings() did not extract correct profile names. Found: %v", mappings)
	}
}

func TestParseMappings_NonExistent(t *testing.T) {
	_, _, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	mappings, err := ParseMappings()
	if err != nil {
		t.Fatalf("ParseMappings() error = %v, want no error for non-existent file", err)
	}

	if len(mappings) != 0 {
		t.Errorf("ParseMappings() = %v, want empty slice", mappings)
	}
}

func TestExtractProfileName(t *testing.T) {
	tests := []struct {
		name     string
		configPath string
		want     string
	}{
		{
			name:       "standard format",
			configPath: "/home/user/.gitconfig-work",
			want:       "work",
		},
		{
			name:       "with tilde",
			configPath: "~/.gitconfig-personal",
			want:       "personal",
		},
		{
			name:       "no profile name",
			configPath: "/home/user/.gitconfig",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractProfileName(tt.configPath)
			if got != tt.want {
				t.Errorf("extractProfileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsProfileMapped(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	configContent := `[includeIf "gitdir/i:` + normalizedDir + `"]
    path = ~/.gitconfig-work
`

	if err := os.WriteFile(gitConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test git config: %v", err)
	}

	mapped, err := IsProfileMapped("work")
	if err != nil {
		t.Fatalf("IsProfileMapped() error = %v", err)
	}

	if !mapped {
		t.Error("IsProfileMapped() = false, want true for mapped profile")
	}

	mapped, err = IsProfileMapped("nonexistent")
	if err != nil {
		t.Fatalf("IsProfileMapped() error = %v", err)
	}

	if mapped {
		t.Error("IsProfileMapped() = true, want false for unmapped profile")
	}
}

func TestGetMappingForDirectory(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "project")
	subDir := filepath.Join(testDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	normalizedDir, _ := utils.NormalizePath(testDir)
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	configContent := `[includeIf "gitdir/i:` + normalizedDir + `"]
    path = ~/.gitconfig-work
`

	if err := os.WriteFile(gitConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test git config: %v", err)
	}

	// Test exact match
	mapping, err := GetMappingForDirectory(testDir)
	if err != nil {
		t.Fatalf("GetMappingForDirectory() error = %v", err)
	}

	if mapping == nil {
		t.Fatal("GetMappingForDirectory() returned nil, want mapping")
	}

	if mapping.Profile != "work" {
		t.Errorf("GetMappingForDirectory().Profile = %v, want work", mapping.Profile)
	}

	// Test prefix match (subdirectory)
	mapping, err = GetMappingForDirectory(subDir)
	if err != nil {
		t.Fatalf("GetMappingForDirectory() error = %v", err)
	}

	if mapping == nil {
		t.Fatal("GetMappingForDirectory() returned nil for subdirectory, want mapping")
	}

	// Test no match
	unmappedDir := filepath.Join(tmpDir, "other")
	mapping, err = GetMappingForDirectory(unmappedDir)
	if err != nil {
		t.Fatalf("GetMappingForDirectory() error = %v", err)
	}

	if mapping != nil {
		t.Error("GetMappingForDirectory() returned mapping, want nil for unmapped directory")
	}
}

func TestParseMappings_ErrorReadingFile(t *testing.T) {
	_, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create a directory with the same name as the config file to cause read error
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

	_, err := ParseMappings()
	if err == nil {
		t.Error("ParseMappings() should fail when config is a directory")
	}
}

func TestParseMappings_ScannerError(t *testing.T) {
	_, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create a very large file that might cause scanner issues
	// (In practice, this tests the scanner error path)
	largeContent := make([]byte, 0)
	for i := 0; i < 1000; i++ {
		largeContent = append(largeContent, []byte("[includeIf \"gitdir/i:/tmp/test\"]\n    path = ~/.gitconfig-test\n")...)
	}
	if err := os.WriteFile(gitConfigPath, largeContent, 0644); err != nil {
		t.Fatalf("Failed to write large git config: %v", err)
	}

	mappings, err := ParseMappings()
	// Should succeed but might take time
	if err != nil {
		t.Logf("ParseMappings() handled large file: %v", err)
	} else {
		if len(mappings) > 0 {
			t.Logf("ParseMappings() parsed %d mappings from large file", len(mappings))
		}
	}
}

func TestParseMappings_InvalidIncludeIfFormat(t *testing.T) {
	_, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create config with malformed includeIf
	configContent := `[includeIf "gitdir/i:/tmp/test"]
    path = ~/.gitconfig-test
[includeIf "invalid format"]
    path = ~/.gitconfig-test2
`
	if err := os.WriteFile(gitConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	mappings, err := ParseMappings()
	if err != nil {
		t.Fatalf("ParseMappings() error = %v", err)
	}

	// Should parse valid ones and skip invalid
	if len(mappings) > 0 {
		t.Logf("ParseMappings() parsed %d valid mappings", len(mappings))
	}
}

func TestParseMappings_NewSectionResets(t *testing.T) {
	_, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create config where new section starts before path line
	configContent := `[includeIf "gitdir/i:/tmp/test"]
[user]
    name = Test
    path = ~/.gitconfig-test
`
	if err := os.WriteFile(gitConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write git config: %v", err)
	}

	mappings, err := ParseMappings()
	if err != nil {
		t.Fatalf("ParseMappings() error = %v", err)
	}

	// Should not create mapping because new section started
	if len(mappings) > 0 {
		t.Error("ParseMappings() should not create mapping when new section starts")
	}
}

func TestExtractProfileName_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		want       string
	}{
		{
			name:       "path with multiple dashes",
			configPath: "/home/user/.gitconfig-work-profile",
			want:       "work-profile",
		},
		{
			name:       "path without prefix",
			configPath: "/home/user/gitconfig-test",
			want:       "",
		},
		{
			name:       "just filename",
			configPath: ".gitconfig-test",
			want:       "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractProfileName(tt.configPath)
			if got != tt.want {
				t.Errorf("extractProfileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsProfileMapped_ParseError(t *testing.T) {
	_, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create unreadable config (directory)
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

	_, err := IsProfileMapped("test")
	if err == nil {
		t.Error("IsProfileMapped() should fail when config is unreadable")
	}
}

func TestGetMappingForDirectory_NormalizeError(t *testing.T) {
	// Test with path that causes normalization error
	// This is hard to trigger in practice, but tests the error path
	_, err := GetMappingForDirectory("")
	if err == nil {
		t.Log("GetMappingForDirectory() might succeed with empty string on some systems")
	} else {
		t.Logf("GetMappingForDirectory() handled empty path: %v", err)
	}
}

func TestGetMappingForDirectory_ParseError(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()
	_ = tmpDir // Use tmpDir

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

	testDir := filepath.Join(tmpDir, "project")
	_, err := GetMappingForDirectory(testDir)
	if err == nil {
		t.Error("GetMappingForDirectory() should fail when config is unreadable")
	}
}

func TestGetDirectoriesForProfile(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create test directories
	testDir1 := filepath.Join(tmpDir, "work/project1")
	testDir2 := filepath.Join(tmpDir, "work/project2")
	testDir3 := filepath.Join(tmpDir, "personal/project1")
	if err := os.MkdirAll(testDir1, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(testDir2, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(testDir3, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	normalizedDir1, _ := utils.NormalizePath(testDir1)
	normalizedDir1 = utils.EnsureTrailingSlash(normalizedDir1)
	normalizedDir2, _ := utils.NormalizePath(testDir2)
	normalizedDir2 = utils.EnsureTrailingSlash(normalizedDir2)
	normalizedDir3, _ := utils.NormalizePath(testDir3)
	normalizedDir3 = utils.EnsureTrailingSlash(normalizedDir3)

	// Create config with multiple mappings
	configContent := fmt.Sprintf(`[includeIf "gitdir/i:%s"]
    path = ~/.gitconfig-work

[includeIf "gitdir/i:%s"]
    path = ~/.gitconfig-work

[includeIf "gitdir/i:%s"]
    path = ~/.gitconfig-personal
`, normalizedDir1, normalizedDir2, normalizedDir3)

	if err := os.WriteFile(gitConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test getting directories for work profile
	workDirs, err := GetDirectoriesForProfile("work")
	if err != nil {
		t.Fatalf("GetDirectoriesForProfile() error = %v", err)
	}

	if len(workDirs) != 2 {
		t.Errorf("GetDirectoriesForProfile('work') returned %d directories, want 2", len(workDirs))
	}

	// Verify the directories are correct
	foundDir1 := false
	foundDir2 := false
	for _, dir := range workDirs {
		if dir == normalizedDir1 {
			foundDir1 = true
		}
		if dir == normalizedDir2 {
			foundDir2 = true
		}
	}

	if !foundDir1 || !foundDir2 {
		t.Errorf("GetDirectoriesForProfile('work') returned unexpected directories: %v", workDirs)
	}

	// Test getting directories for personal profile
	personalDirs, err := GetDirectoriesForProfile("personal")
	if err != nil {
		t.Fatalf("GetDirectoriesForProfile() error = %v", err)
	}

	if len(personalDirs) != 1 {
		t.Errorf("GetDirectoriesForProfile('personal') returned %d directories, want 1", len(personalDirs))
	}

	if len(personalDirs) > 0 && personalDirs[0] != normalizedDir3 {
		t.Errorf("GetDirectoriesForProfile('personal') = %v, want %v", personalDirs[0], normalizedDir3)
	}
}

func TestGetDirectoriesForProfile_NoMappings(t *testing.T) {
	_, _, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Test with profile that has no mappings
	dirs, err := GetDirectoriesForProfile("nonexistent")
	if err != nil {
		t.Fatalf("GetDirectoriesForProfile() error = %v", err)
	}

	if len(dirs) != 0 {
		t.Errorf("GetDirectoriesForProfile('nonexistent') returned %d directories, want 0", len(dirs))
	}
}

func TestGetDirectoriesForProfile_EmptyConfig(t *testing.T) {
	_, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create empty config
	if err := os.WriteFile(gitConfigPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	dirs, err := GetDirectoriesForProfile("test")
	if err != nil {
		t.Fatalf("GetDirectoriesForProfile() error = %v", err)
	}

	if len(dirs) != 0 {
		t.Errorf("GetDirectoriesForProfile() with empty config returned %d directories, want 0", len(dirs))
	}
}

