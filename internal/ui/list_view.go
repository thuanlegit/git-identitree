package ui

import (
	"fmt"
	"strings"

	"github.com/thuanlegit/git-identitree/internal/profile"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62")).
			Padding(1, 0)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			Padding(0, 1)

	rowStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

// ListModel is the Bubble Tea model for listing profiles.
type ListModel struct {
	profiles []profile.Profile
	width    int
	height   int
}

// NewListModel creates a new list model.
func NewListModel(profiles []profile.Profile) *ListModel {
	return &ListModel{
		profiles: profiles,
	}
}

// Init implements the tea.Model interface.
func (m *ListModel) Init() tea.Cmd {
	return nil
}

// Update implements the tea.Model interface.
func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *ListModel) View() string {
	if len(m.profiles) == 0 {
		return titleStyle.Render("No profiles found. Create one with 'gidtree profile create'")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Git Identitree Profiles\n"))
	b.WriteString("\n")

	// Table header
	header := headerStyle.Render(fmt.Sprintf("%-20s %-30s %-30s %-20s %-40s", "Name", "Author Name", "Email", "GPG Key", "SSH Key Path"))
	b.WriteString(header)
	b.WriteString("\n")

	// Table rows
	for _, prof := range m.profiles {
		authorName := prof.GetAuthorName()
		sshKey := prof.SSHKeyPath
		if sshKey == "" {
			sshKey = "(none)"
		}
		gpgKey := prof.GPGKeyID
		if gpgKey == "" {
			gpgKey = "(none)"
		}
		row := rowStyle.Render(fmt.Sprintf("%-20s %-30s %-30s %-20s %-40s", prof.Name, authorName, prof.Email, gpgKey, sshKey))
		b.WriteString(row)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("Press 'q' to quit")

	return b.String()
}

