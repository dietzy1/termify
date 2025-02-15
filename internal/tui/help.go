package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpModel struct {
	keys   KeyMap
	width  int
	height int
}

func newHelp() helpModel {
	return helpModel{
		keys: DefaultKeyMap,
	}
}

func (m helpModel) Init() tea.Cmd {
	return nil
}

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m helpModel) View() string {
	// Styles
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		Bold(true).
		PaddingBottom(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Padding(1, 2)

	// Build help content
	var helpLines []string
	helpLines = append(helpLines, titleStyle.Render("Key Bindings"))

	// Find the longest key length for alignment
	maxKeyLen := 0
	for _, k := range []key.Binding{
		m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right,
		m.keys.Select, m.keys.CycleFocusForward, m.keys.CycleFocusBackward, m.keys.Copy, m.keys.Quit,
	} {
		if len(k.Help().Key) > maxKeyLen {
			maxKeyLen = len(k.Help().Key) + 4
		}
	}

	for _, k := range []struct {
		key  key.Binding
		desc string
	}{
		{m.keys.Up, "Move up"},
		{m.keys.Down, "Move down"},
		{m.keys.Left, "Move left"},
		{m.keys.Right, "Move right"},
		{m.keys.Select, "Select item"},
		{m.keys.CycleFocusForward, "Next focus"},
		{m.keys.CycleFocusBackward, "Previous focus"},
		{m.keys.Copy, "Copy text"},
		{m.keys.Quit, "Exit help"},
	} {
		keyText := fmt.Sprintf("%-*s", maxKeyLen, k.key.Help().Key)
		helpLines = append(helpLines,
			keyStyle.Render(keyText)+"  "+descStyle.Render(k.desc),
		)
	}

	helpContent := borderStyle.Render(lipgloss.JoinVertical(lipgloss.Left, helpLines...))

	// Center help content
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpContent)
}
