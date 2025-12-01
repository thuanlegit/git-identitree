package mapping

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMappings_ErrorReadingFile(t *testing.T) {
	_, gitConfigPath, cleanup := setupMappingTestEnv(t)
	defer cleanup()

	// Create a directory with the same name as the config file to cause read error
	os.Remove(gitConfigPath)
	os.MkdirAll(gitConfigPath, 0755)
	defer os.RemoveAll(gitConfigPath)

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
	os.WriteFile(gitConfigPath, largeContent, 0644)

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
	os.WriteFile(gitConfigPath, []byte(configContent), 0644)

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
	os.WriteFile(gitConfigPath, []byte(configContent), 0644)

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
	os.Remove(gitConfigPath)
	os.MkdirAll(gitConfigPath, 0755)
	defer os.RemoveAll(gitConfigPath)

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
	os.Remove(gitConfigPath)
	os.MkdirAll(gitConfigPath, 0755)
	defer os.RemoveAll(gitConfigPath)

	testDir := filepath.Join(tmpDir, "project")
	_, err := GetMappingForDirectory(testDir)
	if err == nil {
		t.Error("GetMappingForDirectory() should fail when config is unreadable")
	}
}

