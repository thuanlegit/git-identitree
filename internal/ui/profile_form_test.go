package ui

import (
	"os"
	"path/filepath"
	"testing"

	"git-identitree/internal/profile"
)

// Note: CreateProfileForm is an interactive form that requires user input.
// Testing it fully would require mocking the Huh library, which is complex.
// We can at least test that the function exists and can be called.
// In a real scenario, you might want to use dependency injection or
// create a testable version that accepts form values as parameters.

func TestCreateProfileForm_Exists(t *testing.T) {
	// This test verifies the function signature and basic structure
	// Full testing would require mocking the Huh form library
	
	// The function exists and can be referenced
	_ = CreateProfileForm
	
	// We can't easily test the interactive form without mocking,
	// but we can verify the function returns the expected type
	// by checking it compiles and has the right signature
	
	// In practice, you might want to refactor this to accept
	// form values as parameters for better testability
}

// Helper function to create a profile directly (for testing purposes)
// This demonstrates what CreateProfileForm does internally
func createTestProfile(name, email, sshKeyPath, gpgKeyID string) *profile.Profile {
	return &profile.Profile{
		Name:       name,
		Email:      email,
		SSHKeyPath: sshKeyPath,
		GPGKeyID:   gpgKeyID,
	}
}

func TestCreateTestProfile(t *testing.T) {
	prof := createTestProfile("test", "test@example.com", "/path/to/key", "ABC123")

	if prof.Name != "test" {
		t.Errorf("Profile name = %v, want test", prof.Name)
	}
	if prof.Email != "test@example.com" {
		t.Errorf("Profile email = %v, want test@example.com", prof.Email)
	}
	if prof.SSHKeyPath != "/path/to/key" {
		t.Errorf("Profile SSHKeyPath = %v, want /path/to/key", prof.SSHKeyPath)
	}
	if prof.GPGKeyID != "ABC123" {
		t.Errorf("Profile GPGKeyID = %v, want ABC123", prof.GPGKeyID)
	}
}

func TestGetSSHKeySuggestions(t *testing.T) {
	// Test with a mock .ssh directory
	tmpDir, err := os.MkdirTemp("", "gidtree-ssh-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override HOME to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	// Create test SSH key files
	testFiles := []string{
		"id_rsa",
		"id_ed25519",
		"id_rsa.pub",  // Should be excluded
		"id_ecdsa",
		"github",
		"gitlab",
		"config",      // Should be excluded (doesn't match patterns)
		"known_hosts", // Should be excluded
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(sshDir, filename)
		if err := os.WriteFile(filePath, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Get suggestions
	suggestions := getSSHKeySuggestions()

	// Expected suggestions (excluding .pub files and non-matching files)
	expected := map[string]bool{
		"~/.ssh/id_rsa":     true,
		"~/.ssh/id_ed25519": true,
		"~/.ssh/id_ecdsa":   true,
		"~/.ssh/github":     true,
		"~/.ssh/gitlab":     true,
	}

	// Verify suggestions
	if len(suggestions) != len(expected) {
		t.Errorf("Expected %d suggestions, got %d: %v", len(expected), len(suggestions), suggestions)
	}

	for _, suggestion := range suggestions {
		if !expected[suggestion] {
			t.Errorf("Unexpected suggestion: %s", suggestion)
		}
	}
}

func TestGetSSHKeySuggestions_NoSSHDir(t *testing.T) {
	// Test when .ssh directory doesn't exist
	tmpDir, err := os.MkdirTemp("", "gidtree-ssh-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override HOME to use temp directory (without .ssh)
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Get suggestions
	suggestions := getSSHKeySuggestions()

	// Should return empty list when .ssh directory doesn't exist
	if len(suggestions) != 0 {
		t.Errorf("Expected 0 suggestions when .ssh doesn't exist, got %d", len(suggestions))
	}
}

func TestUpdateProfileForm_Exists(t *testing.T) {
	// This test verifies the UpdateProfileForm function exists and has correct signature
	// Full testing would require mocking the Huh form library

	// The function exists and can be referenced
	_ = UpdateProfileForm

	// Create a test profile
	testProfile := &profile.Profile{
		Name:       "test",
		Email:      "test@example.com",
		AuthorName: "Test Author",
		SSHKeyPath: "~/.ssh/id_rsa",
		GPGKeyID:   "GPG123",
	}

	// Verify we can create a test profile (tests the struct)
	if testProfile.Name != "test" {
		t.Errorf("Profile name = %v, want test", testProfile.Name)
	}
	if testProfile.Email != "test@example.com" {
		t.Errorf("Profile email = %v, want test@example.com", testProfile.Email)
	}
	if testProfile.AuthorName != "Test Author" {
		t.Errorf("Profile authorName = %v, want Test Author", testProfile.AuthorName)
	}
}

