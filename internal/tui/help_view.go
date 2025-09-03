package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

func (m applicationModel) renderHelp() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		MarginBottom(1).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(PrimaryColor)

	sectionTitleStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(TextColor)

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, true, true, true).
		BorderForeground(BorderColor).
		Width(m.width - 2).
		Align(lipgloss.Center).
		Height(m.height - lipgloss.Height(m.navbar.View()) - lipgloss.Height(m.playbackControl.View()) - lipgloss.Height(m.audioPlayer.View()) - 3)

	// Group key bindings by category
	navigationBindings := []key.Binding{
		DefaultKeyMap.Up, DefaultKeyMap.Down, DefaultKeyMap.Left, DefaultKeyMap.Right,
		DefaultKeyMap.CycleFocusForward, DefaultKeyMap.CycleFocusBackward,
	}

	actionBindings := []key.Binding{
		DefaultKeyMap.Select, DefaultKeyMap.Copy, DefaultKeyMap.Return, DefaultKeyMap.AddToQueue,
	}

	systemBindings := []key.Binding{
		DefaultKeyMap.Help, DefaultKeyMap.Quit, DefaultKeyMap.Search, DefaultKeyMap.ViewQueue, DefaultKeyMap.DeviceDialog,
	}

	mediaBindings := []key.Binding{
		DefaultKeyMap.VolumeUp, DefaultKeyMap.VolumeDown, DefaultKeyMap.VolumeMute,
		DefaultKeyMap.Previous, DefaultKeyMap.PlayPause, DefaultKeyMap.Next,
		DefaultKeyMap.Shuffle, DefaultKeyMap.Repeat,
	}

	// Function to render a section of key bindings
	renderSection := func(title string, bindings []key.Binding) string {
		var lines []string
		lines = append(lines, sectionTitleStyle.Render(title))

		// Find the longest key length for alignment
		maxKeyLen := 0
		for _, k := range bindings {
			if lipgloss.Width(k.Help().Key) > maxKeyLen {
				maxKeyLen = lipgloss.Width(k.Help().Key)
			}
		}

		// Add each key binding to the section with proper alignment
		for _, k := range bindings {
			keyText := keyStyle.Render(k.Help().Key)

			padding := strings.Repeat(" ", maxKeyLen-lipgloss.Width((k.Help().Key))+4)
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

	return containerStyle.Render(helpContent)
}
