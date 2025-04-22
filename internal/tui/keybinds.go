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
	Select     key.Binding
	Copy       key.Binding
	Return     key.Binding
	AddToQueue key.Binding
	ViewQueue  key.Binding

	// System
	Quit     key.Binding
	Help     key.Binding
	Settings key.Binding
	Search   key.Binding
	Device   key.Binding

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
	Quit: key.NewBinding(
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
		key.WithKeys("enter"),
		key.WithHelp("enter", "select track"),
	),
	AddToQueue: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "Add to queue"),
	),
	ViewQueue: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "View queue"),
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
		key.WithKeys("?"),
		key.WithHelp("?", "show help view"),
	),
	VolumeUp: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("shift+↑", "increase volume"),
	),
	VolumeDown: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift+↓", "lower volume"),
	),
	Settings: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "show settings view"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Device: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "select device"),
	),
	Return: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "return to previous view"),
	),
	Shuffle: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "toggle shuffle"),
	),
	Previous: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "previous song"),
	),
	PlayPause: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "play/pause"),
	),
	Next: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next song"),
	),
	Repeat: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "cycle repeat mode"),
	),
	VolumeMute: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "mute/unmute volume"),
	),
}
