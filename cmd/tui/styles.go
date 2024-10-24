// cmd/tui/styles.go
var (
    subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
    highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
    special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

    titleStyle = lipgloss.NewStyle().
        MarginLeft(2).
        Foreground(highlight)

    selectedItemStyle = lipgloss.NewStyle().
        Foreground(special).
        Bold(true)

    itemStyle = lipgloss.NewStyle().
        Foreground(subtle)

    errorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF0000")).
        Bold(true)
)
