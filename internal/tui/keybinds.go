package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding

	CycleFocusForward  key.Binding
	CycleFocusBackward key.Binding

	//Actions
	Select key.Binding
	Copy   key.Binding
	Return key.Binding

	// System
	QuitApplication key.Binding
	Help            key.Binding
	Settings        key.Binding

	// Media controls
	Shuffle    key.Binding
	Previous   key.Binding
	PlayPause  key.Binding
	Next       key.Binding
	Repeat     key.Binding
	VolumeMute key.Binding
	VolumeUp   key.Binding
	VolumeDown key.Binding
}

var DefaultKeyMap = KeyMap{
	QuitApplication: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit application"),
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

	Help: key.NewBinding(
		key.WithKeys("?", "?"),
		key.WithHelp("?", "show help view"),
	),
	VolumeUp: key.NewBinding(
		key.WithKeys("ctrl+up"),
		key.WithHelp("ctrl+↑", "Increase volume"),
	),
	VolumeDown: key.NewBinding(
		key.WithKeys("ctrl+down"),
		key.WithHelp("ctrl+down↓", "Lower volume"),
	),

	Settings: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "show settings view"),
	),
}
