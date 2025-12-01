package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"git-identitree/internal/profile"
	"git-identitree/internal/utils"
)

// LoadKey adds an SSH key to the SSH agent.
func LoadKey(keyPath string) error {
	// Normalize key path
	normalized, err := utils.NormalizePath(keyPath)
	if err != nil {
		return fmt.Errorf("failed to normalize key path: %w", err)
	}

	// Check if key exists
	if _, err := os.Stat(normalized); os.IsNotExist(err) {
		return fmt.Errorf("SSH key does not exist: %s", normalized)
	}

	// Check if key is already loaded
	loaded, err := CheckKeyLoaded(normalized)
	if err != nil {
		return fmt.Errorf("failed to check if key is loaded: %w", err)
	}
	if loaded {
		return nil // Already loaded
	}

	// Add key to agent
	cmd := exec.Command("ssh-add", normalized)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add SSH key to agent: %w", err)
	}

	return nil
}

// UnloadKey removes an SSH key from the SSH agent.
func UnloadKey(keyPath string) error {
	// Normalize key path
	normalized, err := utils.NormalizePath(keyPath)
	if err != nil {
		return fmt.Errorf("failed to normalize key path: %w", err)
	}

	// Get key fingerprint to identify it in the agent
	cmd := exec.Command("ssh-keygen", "-lf", normalized)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get key fingerprint: %w", err)
	}

	// Extract fingerprint (first field)
	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return fmt.Errorf("unexpected fingerprint format")
	}
	fingerprint := fields[1]

	// Remove key by fingerprint
	cmd = exec.Command("ssh-add", "-d", fingerprint)
	if err := cmd.Run(); err != nil {
		// Try removing by path as fallback
		cmd = exec.Command("ssh-add", "-d", normalized)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to remove SSH key from agent: %w", err)
		}
	}

	return nil
}

// CheckKeyLoaded verifies if an SSH key is loaded in the agent.
func CheckKeyLoaded(keyPath string) (bool, error) {
	// Normalize key path
	normalized, err := utils.NormalizePath(keyPath)
	if err != nil {
		return false, fmt.Errorf("failed to normalize key path: %w", err)
	}

	// Get key fingerprint
	cmd := exec.Command("ssh-keygen", "-lf", normalized)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get key fingerprint: %w", err)
	}

	// Extract fingerprint
	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return false, nil
	}
	fingerprint := fields[1]

	// List keys in agent
	cmd = exec.Command("ssh-add", "-l")
	output, err = cmd.Output()
	if err != nil {
		// SSH agent might not be running
		return false, nil
	}

	// Check if fingerprint is in the list
	return strings.Contains(string(output), fingerprint), nil
}

// LoadKeyForProfile loads the SSH key for a profile if it has one.
func LoadKeyForProfile(prof *profile.Profile) error {
	if prof.SSHKeyPath == "" {
		return nil // No SSH key configured
	}
	return LoadKey(prof.SSHKeyPath)
}

// UnloadKeyForProfile unloads the SSH key for a profile if it has one.
func UnloadKeyForProfile(prof *profile.Profile) error {
	if prof.SSHKeyPath == "" {
		return nil // No SSH key configured
	}
	return UnloadKey(prof.SSHKeyPath)
}

// AutoLoadForDirectory automatically loads the SSH key for the profile mapped to a directory.
func AutoLoadForDirectory(dir string, getProfile func(string) (*profile.Profile, error)) error {
	// This function is a placeholder for future auto-loading functionality.
	// The actual implementation is handled in the CLI layer via the activate command.
	// This function signature might need adjustment based on how it's called.
	return nil
}

