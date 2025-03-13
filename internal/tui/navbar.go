package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

// Dont format
const logo = ` ______              _ ___    
/_  __/__ ______ _  (_) _/_ __
 / / / -_) __/  ' \/ / _/ // /
/_/  \__/_/ /_/_/_/_/_/ \_, / 
                       /___/  `

var _ tea.Model = (*navbarModel)(nil)

type navbarModel struct {
	width int

	deviceSelector DeviceSelectorModel
}

func newNavbar(spotifyState *state.SpotifyState) navbarModel {
	return navbarModel{
		deviceSelector: NewDeviceSelector(spotifyState),
	}
}

func (m navbarModel) Init() tea.Cmd {
	return nil
}

func (m navbarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return m, nil
}

func (m navbarModel) View() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		Bold(true)

	escText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("esc "),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor)).
			Render("Back"),
	)

	helpText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("? "),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor)).
			Render("Help"),
	)

	/* devicesText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("d "),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor)).
			Render("Devices"),
	) */

	/* settings := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("s "),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor)).
			Render("Settings"),
	) */

	rightSection := lipgloss.JoinHorizontal(lipgloss.Right,
		lipgloss.NewStyle().PaddingTop(2).PaddingRight(2).PaddingLeft(2).Render(escText),
		/* settings, */
		lipgloss.NewStyle().PaddingTop(2).PaddingRight(2).PaddingLeft(2).Render(helpText),
		/* lipgloss.NewStyle().MarginTop(2).PaddingRight(2).PaddingLeft(2).Render(devicesText+m.deviceSelector.View()), */
	)

	leftSection := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		MarginLeft(1).
		Render(logo)

	// Create a full-width container with left and right sections
	return lipgloss.NewStyle().
		Width(m.width).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				leftSection,
				lipgloss.NewStyle().Width(m.width-lipgloss.Width(leftSection)-lipgloss.Width(rightSection)).Render(""),
				rightSection,
			),
		)
}
