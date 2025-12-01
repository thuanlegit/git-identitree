package mapping

import (
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
	os.Setenv("HOME", tmpDir)

	gitConfigPath := filepath.Join(tmpDir, ".gitconfig")

	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, gitConfigPath, cleanup
}

func TestParseMappings(t *testing.T) {
	tmpDir, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create a test git config with includeIf blocks
	testDir1 := filepath.Join(tmpDir, "project1")
	testDir2 := filepath.Join(tmpDir, "project2")
	os.MkdirAll(testDir1, 0755)
	os.MkdirAll(testDir2, 0755)

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
	os.MkdirAll(testDir, 0755)
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
	os.MkdirAll(subDir, 0755)
	
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

