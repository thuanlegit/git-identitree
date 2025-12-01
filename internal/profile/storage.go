package profile

import (
	"fmt"
	"os"
	"path/filepath"

	"git-identitree/internal/utils"
	"gopkg.in/yaml.v3"
)

const (
	profilesDir  = ".gidtree"
	profilesFile = "profiles.yaml"
)

// GetProfilesPath returns the path to the profiles.yaml file.
func GetProfilesPath() (string, error) {
	home, err := utils.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, profilesDir, profilesFile), nil
}

// GetProfilesDir returns the path to the .gidtree directory.
func GetProfilesDir() (string, error) {
	home, err := utils.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, profilesDir), nil
}

// LoadProfiles reads and parses the profiles.yaml file.
func LoadProfiles() ([]Profile, error) {
	profilesPath, err := GetProfilesPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
		return []Profile{}, nil
	}

	data, err := os.ReadFile(profilesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles file: %w", err)
	}

	var profiles []Profile
	if err := yaml.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse profiles file: %w", err)
	}

	return profiles, nil
}

// SaveProfiles writes profiles to the profiles.yaml file.
func SaveProfiles(profiles []Profile) error {
	profilesPath, err := GetProfilesPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	profilesDir, err := GetProfilesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	data, err := yaml.Marshal(profiles)
	if err != nil {
		return fmt.Errorf("failed to marshal profiles: %w", err)
	}

	if err := os.WriteFile(profilesPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profiles file: %w", err)
	}

	return nil
}

