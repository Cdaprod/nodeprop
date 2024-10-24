// cmd/nodeprop/tui/views/config.go
type ConfigView struct {
    manager     *nodeprop.NodePropManager
    config      map[string]interface{}
    selected    int
    width       int
    height      int
    loading     bool
    err         error
    editingKey  string
    editingValue string
}

// Main Update function for the TUI
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit
        case key.Matches(msg, m.keys.Help):
            m.help.ShowAll = !m.help.ShowAll
        case key.Matches(msg, m.keys.Tab):
            // Cycle through views
            m.cycleViews()
        }

    case tea.WindowSizeMsg:
        if !m.ready {
            m.viewport = viewport.New(msg.Width, msg.Height-4) // Leave room for help
            m.viewport.SetContent(m.activeView.View())
            m.ready = true
        }
        m.viewport.Width = msg.Width
        m.viewport.Height = msg.Height - 4
        m.activeView.SetSize(msg.Width, msg.Height-4)

    case errMsg:
        m.err = msg.err
        return m, nil
    }

    // Update active view
    newView, cmd := m.activeView.Update(msg)
    m.activeView = newView
    if cmd != nil {
        cmds = append(cmds, cmd)
    }

    return m, tea.Batch(cmds...)
}

// Main View function for the TUI
func (m Model) View() string {
    if !m.ready {
        return "Initializing..."
    }

    // Build the view
    var s strings.Builder

    // Title
    s.WriteString(titleStyle.Render("NodeProp TUI"))
    s.WriteString("\n\n")

    // Main content
    content := m.activeView.View()
    m.viewport.SetContent(content)
    s.WriteString(m.viewport.View())

    // Error message if any
    if m.err != nil {
        s.WriteString("\n")
        s.WriteString(errorStyle.Render(m.err.Error()))
    }

    // Help
    s.WriteString("\n")
    s.WriteString(m.help.View(m.keys))

    return s.String()
}

type errMsg struct {
    err error
}

func (e errMsg) Error() string { return e.err.Error() }