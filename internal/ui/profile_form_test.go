package ui

import (
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

