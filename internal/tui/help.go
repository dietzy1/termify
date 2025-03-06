package tui

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpModel struct {
	width, height int
	keys          KeyMap
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
	case tea.WindowSizeMsg:
		log.Println("Help: Window size message received", msg.Width, msg.Height)
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
		MarginBottom(1).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color(PrimaryColor))

	sectionTitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColor)).
		Width(m.width - 2).Align(lipgloss.Center).Height(m.height - 2)

	// Group key bindings by category
	navigationBindings := []key.Binding{
		m.keys.Up, m.keys.Down, m.keys.Left, m.keys.Right,
		m.keys.CycleFocusForward, m.keys.CycleFocusBackward,
		m.keys.Return,
	}

	actionBindings := []key.Binding{
		m.keys.Select, m.keys.Copy,
	}

	systemBindings := []key.Binding{
		m.keys.Help, m.keys.Quit, m.keys.Settings,
	}

	mediaBindings := []key.Binding{
		m.keys.VolumeUp, m.keys.VolumeDown, m.keys.VolumeMute,
		m.keys.Previous, m.keys.PlayPause, m.keys.Next,
		m.keys.Shuffle, m.keys.Repeat,
	}

	// Function to render a section of key bindings
	renderSection := func(title string, bindings []key.Binding) string {
		var lines []string
		lines = append(lines, sectionTitleStyle.Render(title))

		// Find the longest key length for alignment
		maxKeyLen := 0
		for _, k := range bindings {
			if len(k.Help().Key) > maxKeyLen {
				maxKeyLen = len(k.Help().Key)
			}
		}

		// Add each key binding to the section with proper alignment
		for _, k := range bindings {
			// Add extra padding for better visual spacing
			keyText := keyStyle.Render(k.Help().Key)
			// Pad with spaces to align all descriptions
			padding := strings.Repeat(" ", maxKeyLen-len(k.Help().Key)+4)
			lines = append(lines, keyText+padding+descStyle.Render(k.Help().Desc))
		}

		return lipgloss.JoinVertical(lipgloss.Left, lines...)
	}

	// Build the help content with sections
	navSection := renderSection("Navigation", navigationBindings)
	actionSection := renderSection("Actions", actionBindings)
	systemSection := renderSection("System", systemBindings)
	mediaSection := renderSection("Media Controls", mediaBindings)

	// Determine layout based on available width
	var helpContent string

	// Calculate section widths for horizontal layout
	navWidth := lipgloss.Width(navSection)
	actionWidth := lipgloss.Width(actionSection)
	systemWidth := lipgloss.Width(systemSection)
	mediaWidth := lipgloss.Width(mediaSection)

	// Horizontal spacing between sections
	horizontalPadding := 6

	// Total width needed for horizontal layout
	totalWidth := navWidth + actionWidth + systemWidth + mediaWidth + (horizontalPadding * 3)

	if m.width >= totalWidth {
		// Horizontal layout for wide terminals - all sections in one row
		helpContent = lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("Keyboard Shortcuts"),
			lipgloss.JoinHorizontal(lipgloss.Top,
				navSection,
				lipgloss.NewStyle().PaddingLeft(horizontalPadding).Render(actionSection),
				lipgloss.NewStyle().PaddingLeft(horizontalPadding).Render(systemSection),
				lipgloss.NewStyle().PaddingLeft(horizontalPadding).Render(mediaSection),
			),
		)
	} else if m.width >= (SHRINKWIDTH*3)/2 {
		// Two-row layout for medium terminals
		topRow := lipgloss.JoinHorizontal(lipgloss.Top,
			navSection,
			lipgloss.NewStyle().PaddingLeft(horizontalPadding).Render(actionSection),
		)

		bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
			systemSection,
			lipgloss.NewStyle().PaddingLeft(horizontalPadding).Render(mediaSection),
		)

		helpContent = lipgloss.JoinVertical(lipgloss.Center,
			titleStyle.Render("Keyboard Shortcuts"),
			topRow,
			lipgloss.NewStyle().PaddingTop(1).Render(bottomRow),
		)
	} else {
		// Single-column layout for narrower terminals
		helpContent = lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("Keyboard Shortcuts"),
			navSection,
			lipgloss.NewStyle().PaddingTop(1).Render(actionSection),
			lipgloss.NewStyle().PaddingTop(1).Render(systemSection),
			lipgloss.NewStyle().PaddingTop(1).Render(mediaSection),
		)
	}

	// Apply container style and center in available space
	styledContent := containerStyle.Render(helpContent)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, styledContent)
}
