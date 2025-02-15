package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Quit               key.Binding
	Up                 key.Binding
	Down               key.Binding
	Left               key.Binding
	Right              key.Binding
	Select             key.Binding
	CycleFocusForward  key.Binding
	CycleFocusBackward key.Binding
	Copy               key.Binding
}

var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c/esc", "quit application"),
	),
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("←/h", "previous track"),
	),
	Right: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("→/l", "next track"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter", "space"),
		key.WithHelp("enter/space", "select track"),
	),
	CycleFocusForward: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "cycle focus forward"),
	),
	CycleFocusBackward: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "cycle focus backward"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy URL to clipboard"),
	),
}
