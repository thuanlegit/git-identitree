package mapping

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"git-identitree/internal/profile"
	"git-identitree/internal/utils"
)

// MapProfileToDirectory creates a profile-specific git config and adds an includeIf block.
func MapProfileToDirectory(prof *profile.Profile, dir string) error {
	// Normalize directory path
	normalizedDir, err := utils.NormalizePath(dir)
	if err != nil {
		return fmt.Errorf("failed to normalize directory path: %w", err)
	}
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Check if directory is already mapped
	mappings, err := ParseMappings()
	if err != nil {
		return fmt.Errorf("failed to parse existing mappings: %w", err)
	}
	for _, m := range mappings {
		if m.Directory == normalizedDir {
			return fmt.Errorf("directory '%s' is already mapped to profile '%s'", dir, m.Profile)
		}
	}

	// Generate profile-specific config file
	configPath, err := generateProfileConfig(prof)
	if err != nil {
		return fmt.Errorf("failed to generate profile config: %w", err)
	}

	// Add includeIf block to main git config
	if err := addIncludeIfBlock(normalizedDir, configPath); err != nil {
		return fmt.Errorf("failed to add includeIf block: %w", err)
	}

	return nil
}

// UnmapDirectory removes the includeIf block for a directory.
func UnmapDirectory(dir string) error {
	// Normalize directory path
	normalizedDir, err := utils.NormalizePath(dir)
	if err != nil {
		return fmt.Errorf("failed to normalize directory path: %w", err)
	}
	normalizedDir = utils.EnsureTrailingSlash(normalizedDir)

	// Remove includeIf block
	if err := removeIncludeIfBlock(normalizedDir); err != nil {
		return fmt.Errorf("failed to remove includeIf block: %w", err)
	}

	return nil
}

// generateProfileConfig creates or updates a profile-specific git config file.
func generateProfileConfig(prof *profile.Profile) (string, error) {
	home, err := utils.GetHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(home, fmt.Sprintf(".gitconfig-%s", prof.Name))

	var config strings.Builder
	config.WriteString(fmt.Sprintf("[user]\n"))
	config.WriteString(fmt.Sprintf("    name = %s\n", prof.Name))
	config.WriteString(fmt.Sprintf("    email = %s\n", prof.Email))
	
	if prof.GPGKeyID != "" {
		config.WriteString(fmt.Sprintf("    signingkey = %s\n", prof.GPGKeyID))
	}

	// Configure SSH key if provided
	if prof.SSHKeyPath != "" {
		// Use core.sshCommand to specify the SSH key
		// This approach works with Git's SSH URL rewriting
		config.WriteString(fmt.Sprintf("\n[core]\n"))
		config.WriteString(fmt.Sprintf("    sshCommand = ssh -i %s -F /dev/null\n", prof.SSHKeyPath))
	}

	if err := os.WriteFile(configPath, []byte(config.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write profile config: %w", err)
	}

	return configPath, nil
}

// addIncludeIfBlock adds an includeIf block to ~/.gitconfig.
func addIncludeIfBlock(dir, configPath string) error {
	gitConfigPath, err := getGitConfigPath()
	if err != nil {
		return err
	}

	// Convert configPath to use ~ if it's in home directory
	home, err := utils.GetHomeDir()
	if err == nil && strings.HasPrefix(configPath, home) {
		configPath = strings.Replace(configPath, home, "~", 1)
	}

	// Read existing content
	var lines []string
	if _, err := os.Stat(gitConfigPath); err == nil {
		file, err := os.Open(gitConfigPath)
		if err != nil {
			return fmt.Errorf("failed to open git config: %w", err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		file.Close()
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read git config: %w", err)
		}
	}

	// Check if includeIf block already exists for this directory
	includeIfRegex := regexp.MustCompile(`^\s*\[includeIf\s+"gitdir/i:(.+)"\]\s*$`)
	for i, line := range lines {
		if matches := includeIfRegex.FindStringSubmatch(line); matches != nil {
			existingDir := matches[1]
			normalizedExisting, _ := utils.NormalizePath(existingDir)
			normalizedExisting = utils.EnsureTrailingSlash(normalizedExisting)
			if normalizedExisting == dir {
				// Already exists, update the path line
				if i+1 < len(lines) {
					pathRegex := regexp.MustCompile(`^\s*path\s*=\s*(.+)\s*$`)
					if pathRegex.MatchString(lines[i+1]) {
						lines[i+1] = fmt.Sprintf("    path = %s", configPath)
						// Write back
						return writeGitConfig(gitConfigPath, lines)
					}
				}
			}
		}
	}

	// Append new includeIf block
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf(`[includeIf "gitdir/i:%s"]`, dir))
	lines = append(lines, fmt.Sprintf("    path = %s", configPath))

	return writeGitConfig(gitConfigPath, lines)
}

// removeIncludeIfBlock removes an includeIf block for a directory.
func removeIncludeIfBlock(dir string) error {
	gitConfigPath, err := getGitConfigPath()
	if err != nil {
		return err
	}

	file, err := os.Open(gitConfigPath)
	if err != nil {
		return fmt.Errorf("failed to open git config: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read git config: %w", err)
	}

	includeIfRegex := regexp.MustCompile(`^\s*\[includeIf\s+"gitdir/i:(.+)"\]\s*$`)

	var newLines []string
	var skipNext bool
	for i, line := range lines {
		if skipNext {
			skipNext = false
			continue
		}

		if matches := includeIfRegex.FindStringSubmatch(line); matches != nil {
			existingDir := matches[1]
			normalizedExisting, _ := utils.NormalizePath(existingDir)
			normalizedExisting = utils.EnsureTrailingSlash(normalizedExisting)
			
			if normalizedExisting == dir {
				// Skip this includeIf line and the next path line
				skipNext = true
				// Also skip empty line before if it exists
				if i > 0 && strings.TrimSpace(lines[i-1]) == "" {
					// Remove the last added empty line
					if len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) == "" {
						newLines = newLines[:len(newLines)-1]
					}
				}
				continue
			}
		}

		newLines = append(newLines, line)
	}

	return writeGitConfig(gitConfigPath, newLines)
}

// writeGitConfig writes lines to the git config file.
func writeGitConfig(path string, lines []string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	content := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write git config: %w", err)
	}

	return nil
}

// getGitConfigPath returns the path to ~/.gitconfig.
func getGitConfigPath() (string, error) {
	home, err := utils.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".gitconfig"), nil
}

