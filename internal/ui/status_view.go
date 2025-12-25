package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thuanlegit/git-identitree/internal/mapping"
	"github.com/thuanlegit/git-identitree/internal/profile"
	"github.com/thuanlegit/git-identitree/internal/utils"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	statusTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("62")).
				Padding(1, 0)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Padding(1, 0)

	infoStyle = lipgloss.NewStyle().
			Padding(0, 2)

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	inactiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// StatusModel is the Bubble Tea model for displaying status.
type StatusModel struct {
	mappings      []mapping.Mapping
	currentDir    string
	activeProfile *profile.Profile
	width         int
	height        int
}

// NewStatusModel creates a new status model.
func NewStatusModel() (*StatusModel, error) {
	mappings, err := mapping.ParseMappings()
	if err != nil {
		return nil, err
	}

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = ""
	}

	// Find active profile for current directory
	var activeProfile *profile.Profile
	if currentDir != "" {
		m, err := mapping.GetMappingForDirectory(currentDir)
		if err == nil && m != nil {
			// Load profile
			manager, err := profile.NewManager()
			if err == nil {
				prof, err := manager.GetProfile(m.Profile)
				if err == nil {
					activeProfile = prof
				}
			}
		}
	}

	return &StatusModel{
		mappings:      mappings,
		currentDir:    currentDir,
		activeProfile: activeProfile,
	}, nil
}

// Init implements the tea.Model interface.
func (m *StatusModel) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface.
func (m *StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements the tea.Model interface.
func (m *StatusModel) View() string {
	var b strings.Builder
	b.WriteString(statusTitleStyle.Render("Git Identitree Status\n"))
	b.WriteString("\n")

	// Current directory and active profile
	b.WriteString(sectionStyle.Render("Current Directory"))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(fmt.Sprintf("Path: %s", m.currentDir)))
	b.WriteString("\n\n")

	if m.activeProfile != nil {
		b.WriteString(activeStyle.Render(fmt.Sprintf("✓ Active Profile: %s", m.activeProfile.Name)))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render(fmt.Sprintf("  Email: %s", m.activeProfile.Email)))
		if m.activeProfile.SSHKeyPath != "" {
			b.WriteString("\n")
			b.WriteString(infoStyle.Render(fmt.Sprintf("  SSH Key: %s", m.activeProfile.SSHKeyPath)))
		}
		if m.activeProfile.GPGKeyID != "" {
			b.WriteString("\n")
			b.WriteString(infoStyle.Render(fmt.Sprintf("  GPG Key: %s", m.activeProfile.GPGKeyID)))
		}
	} else {
		b.WriteString(inactiveStyle.Render("No active profile for current directory"))
	}
	b.WriteString("\n\n")

	// Directory mappings
	b.WriteString(sectionStyle.Render("Directory Mappings"))
	b.WriteString("\n")

	if len(m.mappings) == 0 {
		b.WriteString(infoStyle.Render("No directory mappings found."))
		b.WriteString("\n")
	} else {
		for _, m := range m.mappings {
			// Shorten directory path for display
			home, _ := utils.GetHomeDir()
			displayDir := m.Directory
			if strings.HasPrefix(displayDir, home) {
				displayDir = strings.Replace(displayDir, home, "~", 1)
			}
			b.WriteString(infoStyle.Render(fmt.Sprintf("  %s → %s", displayDir, m.Profile)))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// Git config status
	b.WriteString(sectionStyle.Render("Git Config"))
	b.WriteString("\n")
	gitConfigPath, err := getGitConfigPath()
	if err == nil {
		if _, err := os.Stat(gitConfigPath); err == nil {
			b.WriteString(infoStyle.Render(fmt.Sprintf("✓ Main config: %s", gitConfigPath)))
		} else {
			b.WriteString(infoStyle.Render(fmt.Sprintf("✗ Main config not found: %s", gitConfigPath)))
		}
	}
	b.WriteString("\n\n")

	b.WriteString("Press 'q' to quit")

	return b.String()
}

func getGitConfigPath() (string, error) {
	home, err := utils.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gitconfig"), nil
}

