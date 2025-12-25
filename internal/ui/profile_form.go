package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/thuanlegit/git-identitree/internal/profile"
	"github.com/charmbracelet/huh"
)

// getSSHKeySuggestions returns a list of SSH key paths from ~/.ssh directory.
func getSSHKeySuggestions() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return []string{}
	}

	var suggestions []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Filter for common private key files (exclude .pub files)
		if !strings.HasSuffix(name, ".pub") &&
			(strings.HasPrefix(name, "id_") || name == "github" || name == "gitlab" || name == "bitbucket") {
			// Use forward slashes for cross-platform compatibility
			suggestions = append(suggestions, "~/.ssh/"+name)
		}
	}

	return suggestions
}

// CreateProfileForm creates an interactive form for profile creation.
func CreateProfileForm() (*profile.Profile, error) {
	var name, email, authorName, sshKeyPath, gpgKeyID string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Profile Name").
				Description("A unique name for this profile").
				Value(&name).
				Validate(func(s string) error {
					if s == "" {
						return os.ErrInvalid
					}
					return nil
				}),
			huh.NewInput().
				Title("Email").
				Description("Git email address for this profile").
				Value(&email).
				Validate(func(s string) error {
					if s == "" {
						return os.ErrInvalid
					}
					return nil
				}),
			huh.NewInput().
				Title("Author Name").
				Description("Git author name (optional, defaults to profile name)").
				Value(&authorName),
			huh.NewInput().
				Title("SSH Key Path").
				Description("Path to SSH private key (optional)").
				Placeholder("~/.ssh/id_rsa").
				Suggestions(getSSHKeySuggestions()).
				Value(&sshKeyPath),
			huh.NewInput().
				Title("GPG Key ID").
				Description("GPG key ID for signing commits (optional)").
				Value(&gpgKeyID),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	prof := &profile.Profile{
		Name:       name,
		Email:      email,
		AuthorName: authorName,
		SSHKeyPath: sshKeyPath,
		GPGKeyID:   gpgKeyID,
	}

	return prof, nil
}

// UpdateProfileForm creates an interactive form for updating an existing profile.
// The form is pre-populated with the current profile values.
func UpdateProfileForm(currentProfile *profile.Profile) (*profile.Profile, error) {
	// Pre-populate with current values
	name := currentProfile.Name
	email := currentProfile.Email
	authorName := currentProfile.AuthorName
	sshKeyPath := currentProfile.SSHKeyPath
	gpgKeyID := currentProfile.GPGKeyID

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Profile Name").
				Description("A unique name for this profile (cannot be changed)").
				Value(&name).
				Validate(func(s string) error {
					if s == "" {
						return os.ErrInvalid
					}
					// Ensure name hasn't changed
					if s != currentProfile.Name {
						return os.ErrInvalid
					}
					return nil
				}),
			huh.NewInput().
				Title("Email").
				Description("Git email address for this profile").
				Value(&email).
				Validate(func(s string) error {
					if s == "" {
						return os.ErrInvalid
					}
					return nil
				}),
			huh.NewInput().
				Title("Author Name").
				Description("Git author name (optional, defaults to profile name)").
				Value(&authorName),
			huh.NewInput().
				Title("SSH Key Path").
				Description("Path to SSH private key (optional)").
				Placeholder("~/.ssh/id_rsa").
				Suggestions(getSSHKeySuggestions()).
				Value(&sshKeyPath),
			huh.NewInput().
				Title("GPG Key ID").
				Description("GPG key ID for signing commits (optional)").
				Value(&gpgKeyID),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	prof := &profile.Profile{
		Name:       name,
		Email:      email,
		AuthorName: authorName,
		SSHKeyPath: sshKeyPath,
		GPGKeyID:   gpgKeyID,
	}

	return prof, nil
}

