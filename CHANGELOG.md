# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.1] - 2025-12-25

### Added
- Enhanced delete command with automatic directory unmapping
  - Prompts user to unmap all mapped directories before deletion
  - Shows list of mapped directories
  - Auto-unmaps on confirmation with progress messages
- `GetDirectoriesForProfile()` function to retrieve all directories mapped to a profile
- Comprehensive tests for directory mapping and deletion

### Changed
- Delete command now handles mapped profiles gracefully instead of failing
- Updated help text for delete command

## [1.2.0] - 2025-01-25

### Added
- Author name field separate from profile name
  - Optional field that defaults to profile name if not provided
  - Uses `GetAuthorName()` method for fallback logic
- Profile update command (`gidtree profile update <name>`)
  - Interactive form with pre-populated current values
  - Profile name is read-only to prevent mapping conflicts
- Author name column in profile list view
- Shell autocompletion for all commands
  - Bash, Zsh, Fish, PowerShell support
  - Profile name completion for update, delete, and ssh commands
  - Directory path completion for map and unmap commands
- Version command (`gidtree version`)
  - Version can be set at build time with ldflags

### Changed
- Profile list view now displays author name column
- Column widths adjusted for better display
- Git config now uses author name instead of profile name for commits

## [1.1.0] - 2025-01-25

### Added
- SSH key path placeholder (`~/.ssh/id_rsa`) in profile creation form
- SSH key path suggestions from `~/.ssh` directory
  - Automatically detects common key files (id_*, github, gitlab, bitbucket)
  - Filters out .pub files
  - Provides autocompletion in interactive form

### Fixed
- SSH key path validation now properly expands `~` to home directory
  - Added `ExpandPath()` utility function
  - Paths like `~/.ssh/id_rsa_personal` now work correctly
  - Error "SSH key path does not exist" is fixed

### Changed
- Profile creation form now shows SSH key suggestions
- SSH key validation improved with tilde expansion

## [1.0.0] - 2025-01-25

### Added
- Initial release
- Profile management (create, list, delete)
- Directory mapping with Git's includeIf
- SSH key integration
- GPG key support for commit signing
- Interactive TUI for profile creation and listing
- Status command to view mappings
- Profile-to-directory mapping
- Automatic profile activation based on current directory
- SSH key loading/unloading
- Profile validation and safety checks
- Path normalization for cross-platform support
- Profile storage in YAML format (~/.gidtree/profiles.yaml)
- Git config integration with includeIf blocks

### Features
- Create profiles with name, email, SSH key, and GPG key
- Map profiles to directories for automatic switching
- List all profiles in a formatted table
- Delete profiles with safety checks
- Load/unload SSH keys per profile
- Auto-activate profiles based on directory
- Status view showing current mappings

