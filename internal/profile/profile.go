package profile

// Profile represents a Git identity profile.
type Profile struct {
	Name       string `yaml:"name"`
	Email      string `yaml:"email"`
	SSHKeyPath string `yaml:"ssh_key_path,omitempty"`
	GPGKeyID   string `yaml:"gpg_key_id,omitempty"`
}

