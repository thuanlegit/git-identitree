package profile

import (
	"fmt"
	"os"

	"git-identitree/internal/utils"
)

// Manager handles profile CRUD operations.
type Manager struct {
	profiles []Profile
}

// NewManager creates a new profile manager and loads existing profiles.
func NewManager() (*Manager, error) {
	profiles, err := LoadProfiles()
	if err != nil {
		return nil, err
	}
	return &Manager{profiles: profiles}, nil
}

// GetProfile retrieves a profile by name.
func (m *Manager) GetProfile(name string) (*Profile, error) {
	for i := range m.profiles {
		if m.profiles[i].Name == name {
			return &m.profiles[i], nil
		}
	}
	return nil, fmt.Errorf("profile '%s' not found", name)
}

// ListProfiles returns all profiles.
func (m *Manager) ListProfiles() []Profile {
	return m.profiles
}

// AddProfile adds a new profile.
func (m *Manager) AddProfile(profile Profile) error {
	// Check if profile with same name already exists
	for _, p := range m.profiles {
		if p.Name == profile.Name {
			return fmt.Errorf("profile '%s' already exists", profile.Name)
		}
	}

	// Validate SSH key path if provided
	if profile.SSHKeyPath != "" {
		expandedPath, err := utils.ExpandPath(profile.SSHKeyPath)
		if err != nil {
			return fmt.Errorf("failed to expand SSH key path: %w", err)
		}
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			return fmt.Errorf("SSH key path does not exist: %s", profile.SSHKeyPath)
		}
	}

	m.profiles = append(m.profiles, profile)
	return m.save()
}

// UpdateProfile updates an existing profile.
func (m *Manager) UpdateProfile(name string, profile Profile) error {
	for i := range m.profiles {
		if m.profiles[i].Name == name {
			// Validate SSH key path if provided
			if profile.SSHKeyPath != "" {
				expandedPath, err := utils.ExpandPath(profile.SSHKeyPath)
				if err != nil {
					return fmt.Errorf("failed to expand SSH key path: %w", err)
				}
				if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
					return fmt.Errorf("SSH key path does not exist: %s", profile.SSHKeyPath)
				}
			}
			m.profiles[i] = profile
			return m.save()
		}
	}
	return fmt.Errorf("profile '%s' not found", name)
}

// DeleteProfile removes a profile by name.
// It returns an error if the profile is mapped to any directories.
func (m *Manager) DeleteProfile(name string, isMapped func(string) (bool, error)) error {
	// Check if profile exists
	exists := false
	for i := range m.profiles {
		if m.profiles[i].Name == name {
			exists = true
			break
		}
	}
	if !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	// Check if profile is mapped
	if isMapped != nil {
		mapped, err := isMapped(name)
		if err != nil {
			return fmt.Errorf("failed to check profile mappings: %w", err)
		}
		if mapped {
			return fmt.Errorf("profile '%s' is mapped to one or more directories. Please unmap it first", name)
		}
	}

	// Remove profile
	newProfiles := []Profile{}
	for _, p := range m.profiles {
		if p.Name != name {
			newProfiles = append(newProfiles, p)
		}
	}
	m.profiles = newProfiles
	return m.save()
}

// save persists profiles to disk.
func (m *Manager) save() error {
	return SaveProfiles(m.profiles)
}

