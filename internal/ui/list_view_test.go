package ui

import (
	"strings"
	"testing"

	"git-identitree/internal/profile"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewListModel(t *testing.T) {
	profiles := []profile.Profile{
		{Name: "test1", Email: "test1@example.com"},
		{Name: "test2", Email: "test2@example.com"},
	}

	model := NewListModel(profiles)
	if model == nil {
		t.Fatal("NewListModel() returned nil")
	}

	if len(model.profiles) != len(profiles) {
		t.Errorf("NewListModel() profiles count = %d, want %d", len(model.profiles), len(profiles))
	}
}

func TestListModel_Init(t *testing.T) {
	model := NewListModel([]profile.Profile{})
	cmd := model.Init()
	if cmd != nil {
		t.Error("ListModel.Init() should return nil command")
	}
}

func TestListModel_Update_WindowSize(t *testing.T) {
	model := NewListModel([]profile.Profile{})
	
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	updated, cmd := model.Update(msg)
	
	if cmd != nil {
		t.Error("ListModel.Update() should return nil command for WindowSizeMsg")
	}
	
	updatedModel, ok := updated.(*ListModel)
	if !ok {
		t.Fatal("ListModel.Update() returned wrong type")
	}
	
	if updatedModel.width != 80 || updatedModel.height != 24 {
		t.Errorf("ListModel.Update() width/height = %d/%d, want 80/24", updatedModel.width, updatedModel.height)
	}
}

func TestListModel_Update_Quit(t *testing.T) {
	model := NewListModel([]profile.Profile{})
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	updated, cmd := model.Update(msg)
	
	if cmd == nil {
		t.Error("ListModel.Update() should return quit command for 'q' key")
	}
	
	// Check that cmd is a quit command
	_, ok := updated.(*ListModel)
	if !ok {
		t.Fatal("ListModel.Update() returned wrong type")
	}
}

func TestListModel_View_Empty(t *testing.T) {
	model := NewListModel([]profile.Profile{})
	view := model.View()
	
	if !strings.Contains(view, "No profiles found") {
		t.Error("ListModel.View() should show message for empty profiles")
	}
}

func TestListModel_View_WithProfiles(t *testing.T) {
	profiles := []profile.Profile{
		{Name: "test1", Email: "test1@example.com", SSHKeyPath: "/path/to/key1"},
		{Name: "test2", Email: "test2@example.com"},
	}

	model := NewListModel(profiles)
	view := model.View()
	
	if !strings.Contains(view, "test1") {
		t.Error("ListModel.View() should contain profile name")
	}
	if !strings.Contains(view, "test1@example.com") {
		t.Error("ListModel.View() should contain profile email")
	}
	if !strings.Contains(view, "/path/to/key1") {
		t.Error("ListModel.View() should contain SSH key path")
	}
	if !strings.Contains(view, "(none)") {
		t.Error("ListModel.View() should show '(none)' for profiles without SSH key")
	}
}

