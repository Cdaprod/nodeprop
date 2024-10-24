// cmd/tui/keys.go
type keyMap struct {
    Up        key.Binding
    Down      key.Binding
    Left      key.Binding
    Right     key.Binding
    Help      key.Binding
    Quit      key.Binding
    Tab       key.Binding
    Enter     key.Binding
    Back      key.Binding
}

func newKeyMap() keyMap {
    return keyMap{
        Up: key.NewBinding(
            key.WithKeys("up", "k"),
            key.WithHelp("↑/k", "up"),
        ),
        Down: key.NewBinding(
            key.WithKeys("down", "j"),
            key.WithHelp("↓/j", "down"),
        ),
        Left: key.NewBinding(
            key.WithKeys("left", "h"),
            key.WithHelp("←/h", "previous"),
        ),
        Right: key.NewBinding(
            key.WithKeys("right", "l"),
            key.WithHelp("→/l", "next"),
        ),
        Help: key.NewBinding(
            key.WithKeys("?"),
            key.WithHelp("?", "toggle help"),
        ),
        Quit: key.NewBinding(
            key.WithKeys("q", "ctrl+c"),
            key.WithHelp("q", "quit"),
        ),
        Tab: key.NewBinding(
            key.WithKeys("tab"),
            key.WithHelp("tab", "next view"),
        ),
        Enter: key.NewBinding(
            key.WithKeys("enter"),
            key.WithHelp("enter", "select"),
        ),
        Back: key.NewBinding(
            key.WithKeys("esc"),
            key.WithHelp("esc", "back"),
        ),
    }
}