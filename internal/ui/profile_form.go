package ui

import (
	"os"

	"git-identitree/internal/profile"
	"github.com/charmbracelet/huh"
)

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

