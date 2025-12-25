package mapping

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"git-identitree/internal/utils"
)

// Mapping represents a directory-to-profile mapping.
type Mapping struct {
	Directory string
	Profile   string
	ConfigPath string
}

// ParseMappings extracts all directory-to-profile mappings from ~/.gitconfig.
func ParseMappings() ([]Mapping, error) {
	gitConfigPath, err := getGitConfigPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(gitConfigPath); os.IsNotExist(err) {
		return []Mapping{}, nil
	}

	file, err := os.Open(gitConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git config: %w", err)
	}
	defer file.Close()

	var mappings []Mapping
	scanner := bufio.NewScanner(file)
	
	// Regex to match includeIf blocks
	// [includeIf "gitdir/i:/path/to/dir/"]
	includeIfRegex := regexp.MustCompile(`^\s*\[includeIf\s+"gitdir/i:(.+)"\]\s*$`)
	pathRegex := regexp.MustCompile(`^\s*path\s*=\s*(.+)\s*$`)

	var currentDir string
	var inIncludeIfBlock bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Check for includeIf block
		if matches := includeIfRegex.FindStringSubmatch(line); matches != nil {
			dir := matches[1]
			// Normalize the directory path
			normalized, err := utils.NormalizePath(dir)
			if err != nil {
				// If normalization fails, use original
				normalized = dir
			}
			currentDir = utils.EnsureTrailingSlash(normalized)
			inIncludeIfBlock = true
			continue
		}

		// Check for path line within includeIf block
		if inIncludeIfBlock {
			if matches := pathRegex.FindStringSubmatch(line); matches != nil {
				configPath := strings.TrimSpace(matches[1])
				// Expand ~ in config path
				if strings.HasPrefix(configPath, "~") {
					home, err := utils.GetHomeDir()
					if err == nil {
						configPath = strings.Replace(configPath, "~", home, 1)
					}
				}
				
				// Extract profile name from config path
				// ~/.gitconfig-${profile_name}
				profileName := extractProfileName(configPath)
				
				mappings = append(mappings, Mapping{
					Directory:  currentDir,
					Profile:    profileName,
					ConfigPath: configPath,
				})
				inIncludeIfBlock = false
				currentDir = ""
			} else if strings.HasPrefix(line, "[") {
				// New section started, reset
				inIncludeIfBlock = false
				currentDir = ""
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan git config: %w", err)
	}

	return mappings, nil
}

// extractProfileName extracts the profile name from a config path like ~/.gitconfig-${profile_name}.
func extractProfileName(configPath string) string {
	base := filepath.Base(configPath)
	if strings.HasPrefix(base, ".gitconfig-") {
		return strings.TrimPrefix(base, ".gitconfig-")
	}
	return ""
}

// IsProfileMapped checks if a profile is mapped to any directory.
func IsProfileMapped(profileName string) (bool, error) {
	mappings, err := ParseMappings()
	if err != nil {
		return false, err
	}

	for _, m := range mappings {
		if m.Profile == profileName {
			return true, nil
		}
	}
	return false, nil
}

// GetMappingForDirectory returns the mapping for a given directory, if any.
func GetMappingForDirectory(dir string) (*Mapping, error) {
	normalized, err := utils.NormalizePath(dir)
	if err != nil {
		return nil, err
	}
	normalized = utils.EnsureTrailingSlash(normalized)

	mappings, err := ParseMappings()
	if err != nil {
		return nil, err
	}

	// Check for exact match first
	for _, m := range mappings {
		if m.Directory == normalized {
			return &m, nil
		}
	}

	// Check for prefix match (directory is within mapped directory)
	for _, m := range mappings {
		if strings.HasPrefix(normalized, m.Directory) {
			return &m, nil
		}
	}

	return nil, nil
}

// GetDirectoriesForProfile returns all directories mapped to a specific profile.
func GetDirectoriesForProfile(profileName string) ([]string, error) {
	mappings, err := ParseMappings()
	if err != nil {
		return nil, err
	}

	var directories []string
	for _, m := range mappings {
		if m.Profile == profileName {
			directories = append(directories, m.Directory)
		}
	}

	return directories, nil
}

