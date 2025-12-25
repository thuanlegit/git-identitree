package profile

// Profile represents a Git identity profile.
type Profile struct {
	Name       string `yaml:"name"`
	Email      string `yaml:"email"`
	AuthorName string `yaml:"author_name,omitempty"`
	SSHKeyPath string `yaml:"ssh_key_path,omitempty"`
	GPGKeyID   string `yaml:"gpg_key_id,omitempty"`
}

// GetAuthorName returns the author name, falling back to the profile name if not set.
func (p *Profile) GetAuthorName() string {
	if p.AuthorName != "" {
		return p.AuthorName
	}
	return p.Name
}

