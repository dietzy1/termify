package tui

import (
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

const SHRINKHEIGHT = 24

type navbarModel struct {
	width, applicationHeight int
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
		m.applicationHeight = msg.Height
	}
	return m, nil
}

func (m navbarModel) View() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		Bold(true)

	helpText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("? "),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor)).
			Render("Help"),
	)

	devicesText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("d "),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(TextColor)).
			Render("Devices"),
	)

	var paddingTop = 0
	if m.applicationHeight > SHRINKHEIGHT {
		paddingTop = 2
	}

	rightSection := lipgloss.JoinHorizontal(lipgloss.Right,
		lipgloss.NewStyle().
			PaddingTop(paddingTop).
			PaddingRight(2).
			PaddingLeft(2).
			Render(devicesText),
		lipgloss.NewStyle().
			PaddingTop(paddingTop).
			PaddingRight(2).
			PaddingLeft(2).
			Render(helpText),
	)

	leftSection := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		MarginLeft(1).
		Render(func() string {
			if m.applicationHeight > SHRINKHEIGHT {
				return logo
			}
			return smallLogo
		}())

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
