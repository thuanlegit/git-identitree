package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-identitree/internal/mapping"
	"git-identitree/internal/profile"
	tea "github.com/charmbracelet/bubbletea"
)

func setupStatusTestEnv(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "gidtree-status-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestNewStatusModel(t *testing.T) {
	tmpDir, cleanup := setupStatusTestEnv(t)
	defer cleanup()

	// Initialize profiles
	profilesDir, _ := profile.GetProfilesDir()
	os.MkdirAll(profilesDir, 0755)

	// Create a profile
	manager, err := profile.NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	testProfile := profile.Profile{
		Name:  "test",
		Email: "test@example.com",
	}
	if err := manager.AddProfile(testProfile); err != nil {
		t.Fatalf("AddProfile() error = %v", err)
	}

	// Create test directory and map it
	testDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(testDir, 0755)

	prof, err := manager.GetProfile("test")
	if err != nil {
		t.Fatalf("GetProfile() error = %v", err)
	}

	if err := mapping.MapProfileToDirectory(prof, testDir); err != nil {
		t.Fatalf("MapProfileToDirectory() error = %v", err)
	}

	// Change to test directory
	originalDir, _ := os.Getwd()
	os.Chdir(testDir)
	defer os.Chdir(originalDir)

	// Create status model
	model, err := NewStatusModel()
	if err != nil {
		t.Fatalf("NewStatusModel() error = %v", err)
	}

	if model == nil {
		t.Fatal("NewStatusModel() returned nil")
	}

	if model.activeProfile == nil {
		t.Error("NewStatusModel() should find active profile for mapped directory")
	}

	if len(model.mappings) == 0 {
		t.Error("NewStatusModel() should find mappings")
	}
}

func TestNewStatusModel_NoMapping(t *testing.T) {
	_, cleanup := setupStatusTestEnv(t)
	defer cleanup()

	// Create unmapped directory
	testDir, err := os.MkdirTemp("", "unmapped-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	originalDir, _ := os.Getwd()
	os.Chdir(testDir)
	defer os.Chdir(originalDir)

	model, err := NewStatusModel()
	if err != nil {
		t.Fatalf("NewStatusModel() error = %v", err)
	}

	if model.activeProfile != nil {
		t.Error("NewStatusModel() should not find active profile for unmapped directory")
	}
}

func TestStatusModel_Init(t *testing.T) {
	model := &StatusModel{}
	cmd := model.Init()
	if cmd != nil {
		t.Error("StatusModel.Init() should return nil command")
	}
}

func TestStatusModel_Update_WindowSize(t *testing.T) {
	model := &StatusModel{}
	
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updated, cmd := model.Update(msg)
	
	if cmd != nil {
		t.Error("StatusModel.Update() should return nil command for WindowSizeMsg")
	}
	
	updatedModel, ok := updated.(*StatusModel)
	if !ok {
		t.Fatal("StatusModel.Update() returned wrong type")
	}
	
	if updatedModel.width != 100 || updatedModel.height != 50 {
		t.Errorf("StatusModel.Update() width/height = %d/%d, want 100/50", updatedModel.width, updatedModel.height)
	}
}

func TestStatusModel_Update_Quit(t *testing.T) {
	model := &StatusModel{}
	
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	updated, cmd := model.Update(msg)
	
	if cmd == nil {
		t.Error("StatusModel.Update() should return quit command for 'q' key")
	}
	
	_, ok := updated.(*StatusModel)
	if !ok {
		t.Fatal("StatusModel.Update() returned wrong type")
	}
}

func TestStatusModel_View_WithActiveProfile(t *testing.T) {
	tmpDir, cleanup := setupStatusTestEnv(t)
	defer cleanup()

	model := &StatusModel{
		currentDir: tmpDir,
		activeProfile: &profile.Profile{
			Name:       "test",
			Email:      "test@example.com",
			SSHKeyPath: "/path/to/key",
			GPGKeyID:   "ABC123",
		},
		mappings: []mapping.Mapping{
			{Directory: tmpDir + "/", Profile: "test"},
		},
	}

	view := model.View()
	
	if !strings.Contains(view, "test") {
		t.Error("StatusModel.View() should contain profile name")
	}
	if !strings.Contains(view, "test@example.com") {
		t.Error("StatusModel.View() should contain email")
	}
	if !strings.Contains(view, "/path/to/key") {
		t.Error("StatusModel.View() should contain SSH key path")
	}
	if !strings.Contains(view, "ABC123") {
		t.Error("StatusModel.View() should contain GPG key ID")
	}
	if !strings.Contains(view, "Active Profile") {
		t.Error("StatusModel.View() should show active profile")
	}
}

func TestStatusModel_View_NoActiveProfile(t *testing.T) {
	model := &StatusModel{
		currentDir: "/some/dir",
		activeProfile: nil,
		mappings: []mapping.Mapping{},
	}

	view := model.View()
	
	if !strings.Contains(view, "No active profile") {
		t.Error("StatusModel.View() should show message when no active profile")
	}
}

func TestStatusModel_View_WithMappings(t *testing.T) {
	tmpDir, cleanup := setupStatusTestEnv(t)
	defer cleanup()

	model := &StatusModel{
		mappings: []mapping.Mapping{
			{Directory: tmpDir + "/project1/", Profile: "work"},
			{Directory: tmpDir + "/project2/", Profile: "personal"},
		},
	}

	view := model.View()
	
	if !strings.Contains(view, "work") {
		t.Error("StatusModel.View() should contain mapping profile")
	}
	if !strings.Contains(view, "personal") {
		t.Error("StatusModel.View() should contain all mappings")
	}
}

func TestStatusModel_View_GitConfigExists(t *testing.T) {
	tmpDir, cleanup := setupStatusTestEnv(t)
	defer cleanup()

	// Create git config
	gitConfigPath := filepath.Join(tmpDir, ".gitconfig")
	os.WriteFile(gitConfigPath, []byte("[user]\n"), 0644)

	model := &StatusModel{}
	view := model.View()
	
	if !strings.Contains(view, "Main config") {
		t.Error("StatusModel.View() should show git config status")
	}
}

func TestGetGitConfigPath(t *testing.T) {
	tmpDir, cleanup := setupStatusTestEnv(t)
	defer cleanup()

	path, err := getGitConfigPath()
	if err != nil {
		t.Fatalf("getGitConfigPath() error = %v", err)
	}

	expected := filepath.Join(tmpDir, ".gitconfig")
	if path != expected {
		t.Errorf("getGitConfigPath() = %v, want %v", path, expected)
	}
}

