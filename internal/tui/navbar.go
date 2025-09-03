package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Dont format
const logo = ` ______              _ ___    
/_  __/__ ______ _  (_) _/_ __
 / / / -_) __/  ' \/ / _/ // /
/_/  \__/_/ /_/_/_/_/_/ \_, / 
                       /___/  `

const smallLogo = `Termify`

var _ tea.Model = (*navbarModel)(nil)

type navbarModel struct {
	width, height int
	deviceCount   int
	queueCount    int
}

func newNavbar() navbarModel {
	return navbarModel{}
}

func (m navbarModel) Init() tea.Cmd {
	return nil
}

func (m navbarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m navbarModel) View() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	helpText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("? "),
		lipgloss.NewStyle().
			Foreground(TextColor).
			Render("Help"),
	)

	deviceText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render(DefaultKeyMap.DeviceDialog.Keys()...),
		lipgloss.NewStyle().
			Foreground(TextColor).
			MarginLeft(1).
			Render(fmt.Sprintf("Devices (%d)", m.deviceCount)),
	)

	queueText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render(DefaultKeyMap.ViewQueue.Keys()...),
		lipgloss.NewStyle().
			Foreground(TextColor).
			MarginLeft(1).
			Render(fmt.Sprintf("Queue (%d)", m.queueCount)),
	)

	var paddingTop = 0
	if m.height > 1 {
		paddingTop = 2
	}

	rightSection := lipgloss.JoinHorizontal(lipgloss.Right,
		lipgloss.NewStyle().
			PaddingTop(paddingTop).
			PaddingRight(2).
			PaddingLeft(2).
			Render(queueText),
		lipgloss.NewStyle().
			PaddingTop(paddingTop).
			PaddingRight(2).
			PaddingLeft(2).
			Render(deviceText),
		lipgloss.NewStyle().
			PaddingTop(paddingTop).
			PaddingRight(2).
			PaddingLeft(2).
			Render(helpText),
	)

	leftSection := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		MarginLeft(1).
		Render(func() string {
			if m.height > 1 {
				return logo
			}
			return smallLogo
		}())

	availableWidth := max(0, m.width-lipgloss.Width(leftSection)-lipgloss.Width(rightSection))

	// Create a full-width container with left and right sections
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				leftSection,
				lipgloss.NewStyle().Width(availableWidth).Render(""),
				rightSection,
			),
		)
}
