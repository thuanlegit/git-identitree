package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/thuanlegit/git-identitree/internal/mapping"
	"github.com/thuanlegit/git-identitree/internal/profile"
	"github.com/thuanlegit/git-identitree/internal/ssh"
	"github.com/thuanlegit/git-identitree/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// version can be set at build time using -ldflags "-X main.version=x.y.z"
var version = "1.2.1"

var rootCmd = &cobra.Command{
	Use:   "gidtree",
	Short: "Git Identitree - Manage Git profiles with directory-based context switching",
	Long:  "A CLI tool to manage multiple Git identities and automatically switch between them based on directory context.",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Git Identitree",
	Long:  "Create the necessary working directory (~/.gidtree/) and ensure permissions are correct",
	RunE: func(cmd *cobra.Command, args []string) error {
		profilesDir, err := profile.GetProfilesDir()
		if err != nil {
			return fmt.Errorf("failed to get profiles directory: %w", err)
		}

		if err := os.MkdirAll(profilesDir, 0755); err != nil {
			return fmt.Errorf("failed to create profiles directory: %w", err)
		}

		profilesPath, err := profile.GetProfilesPath()
		if err != nil {
			return fmt.Errorf("failed to get profiles path: %w", err)
		}

		// Create empty profiles file if it doesn't exist
		if _, err := os.Stat(profilesPath); os.IsNotExist(err) {
			if err := profile.SaveProfiles([]profile.Profile{}); err != nil {
				return fmt.Errorf("failed to create profiles file: %w", err)
			}
		}

		fmt.Printf("✓ Initialized Git Identitree at %s\n", profilesDir)
		return nil
	},
}

var profileCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile",
	Long:  "Interactively create a new Git profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		prof, err := ui.CreateProfileForm()
		if err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		manager, err := profile.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize profile manager: %w", err)
		}

		if err := manager.AddProfile(*prof); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}

		fmt.Printf("✓ Profile '%s' created successfully\n", prof.Name)
		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	Long:  "Display all stored profiles with their core settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := profile.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize profile manager: %w", err)
		}

		profiles := manager.ListProfiles()
		model := ui.NewListModel(profiles)

		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run UI: %w", err)
		}

		return nil
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a profile",
	Long:  "Delete a profile. If mapped to directories, will prompt to unmap them first.",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		manager, err := profile.NewManager()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		profiles := manager.ListProfiles()
		var names []string
		for _, p := range profiles {
			names = append(names, p.Name)
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		manager, err := profile.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize profile manager: %w", err)
		}

		// Check if profile exists
		_, err = manager.GetProfile(profileName)
		if err != nil {
			return fmt.Errorf("profile not found: %w", err)
		}

		// Get all directories mapped to this profile
		directories, err := mapping.GetDirectoriesForProfile(profileName)
		if err != nil {
			return fmt.Errorf("failed to check profile mappings: %w", err)
		}

		// If profile is mapped, ask user if they want to unmap
		if len(directories) > 0 {
			fmt.Printf("Profile '%s' is mapped to the following directories:\n", profileName)
			for _, dir := range directories {
				fmt.Printf("  - %s\n", dir)
			}
			fmt.Print("\nDo you want to unmap all directories and delete the profile? (y/N): ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Delete cancelled.")
				return nil
			}

			// Unmap all directories
			fmt.Println("\nUnmapping directories...")
			for _, dir := range directories {
				if err := mapping.UnmapDirectory(dir); err != nil {
					return fmt.Errorf("failed to unmap directory '%s': %w", dir, err)
				}
				fmt.Printf("  ✓ Unmapped: %s\n", dir)
			}
		}

		// Delete the profile (no need to check mappings again)
		isMapped := func(name string) (bool, error) {
			return false, nil // Already handled above
		}

		if err := manager.DeleteProfile(profileName, isMapped); err != nil {
			return fmt.Errorf("failed to delete profile: %w", err)
		}

		fmt.Printf("\n✓ Profile '%s' deleted successfully\n", profileName)
		return nil
	},
}

var profileUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update an existing profile",
	Long:  "Interactively update an existing Git profile with pre-populated values",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		manager, err := profile.NewManager()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		profiles := manager.ListProfiles()
		var names []string
		for _, p := range profiles {
			names = append(names, p.Name)
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		manager, err := profile.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize profile manager: %w", err)
		}

		// Get the current profile
		currentProfile, err := manager.GetProfile(profileName)
		if err != nil {
			return fmt.Errorf("profile not found: %w", err)
		}

		// Show update form with pre-populated values
		updatedProfile, err := ui.UpdateProfileForm(currentProfile)
		if err != nil {
			return fmt.Errorf("failed to update profile: %w", err)
		}

		// Update the profile
		if err := manager.UpdateProfile(profileName, *updatedProfile); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}

		fmt.Printf("✓ Profile '%s' updated successfully\n", profileName)
		return nil
	},
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles",
	Long:  "Commands for managing Git profiles",
}

var mapCmd = &cobra.Command{
	Use:   "map [profile] [directory]",
	Short: "Map a profile to a directory",
	Long:  "Associate a profile with a target directory path. Git will automatically use this profile when working in that directory.",
	Args:  cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: profile name - get list of profiles
			manager, err := profile.NewManager()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			profiles := manager.ListProfiles()
			var names []string
			for _, p := range profiles {
				names = append(names, p.Name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		} else if len(args) == 1 {
			// Second argument: directory path - enable directory completion
			return nil, cobra.ShellCompDirectiveFilterDirs
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		dir := args[1]

		manager, err := profile.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize profile manager: %w", err)
		}

		prof, err := manager.GetProfile(profileName)
		if err != nil {
			return fmt.Errorf("profile not found: %w", err)
		}

		if err := mapping.MapProfileToDirectory(prof, dir); err != nil {
			return fmt.Errorf("failed to map profile: %w", err)
		}

		fmt.Printf("✓ Profile '%s' mapped to directory '%s'\n", profileName, dir)
		return nil
	},
}

var unmapCmd = &cobra.Command{
	Use:   "unmap [directory]",
	Short: "Remove a directory mapping",
	Long:  "Remove the association between a directory and its profile",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Enable directory completion
		return nil, cobra.ShellCompDirectiveFilterDirs
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		if err := mapping.UnmapDirectory(dir); err != nil {
			return fmt.Errorf("failed to unmap directory: %w", err)
		}

		fmt.Printf("✓ Directory '%s' unmapped successfully\n", dir)
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status and mappings",
	Long:  "Display which directories are mapped to which profiles and verify the ~/.gitconfig file",
	RunE: func(cmd *cobra.Command, args []string) error {
		model, err := ui.NewStatusModel()
		if err != nil {
			return fmt.Errorf("failed to create status model: %w", err)
		}

		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("failed to run UI: %w", err)
		}

		return nil
	},
}

var sshLoadCmd = &cobra.Command{
	Use:   "load [profile]",
	Short: "Load SSH key for a profile",
	Long:  "Manually load the SSH key associated with a profile into the SSH agent",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		manager, err := profile.NewManager()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		profiles := manager.ListProfiles()
		var names []string
		for _, p := range profiles {
			if p.SSHKeyPath != "" {
				names = append(names, p.Name)
			}
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		manager, err := profile.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize profile manager: %w", err)
		}

		prof, err := manager.GetProfile(profileName)
		if err != nil {
			return fmt.Errorf("profile not found: %w", err)
		}

		if prof.SSHKeyPath == "" {
			return fmt.Errorf("profile '%s' does not have an SSH key configured", profileName)
		}

		if err := ssh.LoadKeyForProfile(prof); err != nil {
			return fmt.Errorf("failed to load SSH key: %w", err)
		}

		fmt.Printf("✓ SSH key loaded for profile '%s'\n", profileName)
		return nil
	},
}

var sshUnloadCmd = &cobra.Command{
	Use:   "unload [profile]",
	Short: "Unload SSH key for a profile",
	Long:  "Manually unload the SSH key associated with a profile from the SSH agent",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		manager, err := profile.NewManager()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		profiles := manager.ListProfiles()
		var names []string
		for _, p := range profiles {
			if p.SSHKeyPath != "" {
				names = append(names, p.Name)
			}
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]

		manager, err := profile.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize profile manager: %w", err)
		}

		prof, err := manager.GetProfile(profileName)
		if err != nil {
			return fmt.Errorf("profile not found: %w", err)
		}

		if prof.SSHKeyPath == "" {
			return fmt.Errorf("profile '%s' does not have an SSH key configured", profileName)
		}

		if err := ssh.UnloadKeyForProfile(prof); err != nil {
			return fmt.Errorf("failed to unload SSH key: %w", err)
		}

		fmt.Printf("✓ SSH key unloaded for profile '%s'\n", profileName)
		return nil
	},
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Manage SSH keys",
	Long:  "Commands for managing SSH keys in the SSH agent",
}

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Auto-detect and activate profile for current directory",
	Long:  "Automatically detect the current directory, find its mapped profile, and load the associated SSH key if needed",
	RunE: func(cmd *cobra.Command, args []string) error {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		m, err := mapping.GetMappingForDirectory(currentDir)
		if err != nil {
			return fmt.Errorf("failed to get mapping: %w", err)
		}

		if m == nil {
			fmt.Println("No profile mapped for current directory")
			return nil
		}

		manager, err := profile.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize profile manager: %w", err)
		}

		prof, err := manager.GetProfile(m.Profile)
		if err != nil {
			return fmt.Errorf("profile not found: %w", err)
		}

		fmt.Printf("Active profile: %s\n", prof.Name)
		fmt.Printf("Email: %s\n", prof.Email)

		if prof.SSHKeyPath != "" {
			if err := ssh.LoadKeyForProfile(prof); err != nil {
				return fmt.Errorf("failed to load SSH key: %w", err)
			}
			fmt.Printf("✓ SSH key loaded\n")
		}

		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the version of gidtree",
	Long:  "Display the current version of the Git Identitree CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gidtree version %s\n", version)
	},
}

func init() {
	// Profile subcommands
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileUpdateCmd)
	profileCmd.AddCommand(profileDeleteCmd)

	// SSH subcommands
	sshCmd.AddCommand(sshLoadCmd)
	sshCmd.AddCommand(sshUnloadCmd)

	// Root commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(mapCmd)
	rootCmd.AddCommand(unmapCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(sshCmd)
	rootCmd.AddCommand(activateCmd)
	rootCmd.AddCommand(versionCmd)

	// Enable shell completion
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
