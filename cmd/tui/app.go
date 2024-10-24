// cmd/tui/app.go
package tui

import (
    "github.com/charmbracelet/bubbles/help"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/Cdaprod/nodeprop/pkg/nodeprop"
)

// Model represents the TUI state
type Model struct {
    keys       keyMap
    help       help.Model
    viewport   viewport.Model
    manager    *nodeprop.NodePropManager
    activeView View
    views      map[string]View
    ready      bool
    err        error
}

// View interface for different screens
type View interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (View, tea.Cmd)
    View() string
    SetSize(width, height int)
}

// Initialize the TUI
func NewModel(manager *nodeprop.NodePropManager) Model {
    m := Model{
        keys:     newKeyMap(),
        help:     help.New(),
        manager:  manager,
        views:    make(map[string]View),
    }

    // Initialize views
    m.views["workflows"] = NewWorkflowsView(manager)
    m.views["secrets"] = NewSecretsView(manager)
    m.views["files"] = NewFilesView(manager)
    m.views["config"] = NewConfigView(manager)
    
    m.activeView = m.views["workflows"]

    return m
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.activeView.Init(),
        tea.EnterAltScreen,
    )
}