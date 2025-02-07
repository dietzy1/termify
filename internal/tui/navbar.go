package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const logo = `
 ______              _ ___    
/_  __/__ ______ _  (_) _/_ __
 / / / -_) __/  ' \/ / _/ // /
/_/  \__/_/ /_/_/_/_/_/ \_, / 
                       /___/  
`

var _ tea.Model = (*navbarModel)(nil)

type navbarModel struct {
	width int

	//textinput textinput.Model
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
	}
	return m, nil
}

func (m navbarModel) View() string {

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		Align(lipgloss.Center).Width(m.width).Background(lipgloss.Color(BackgroundColor)).PaddingRight(2)

	/* searchBarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColor))
		//Width(m.width)

	questionMarkStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColor))

	containerStyle := lipgloss.NewStyle().
		Width(m.width)

	joined := lipgloss.JoinHorizontal(
		lipgloss.Center,
		searchBarStyle.Render(m.textinput.View()),
		headerStyle.Render(logo),
		questionMarkStyle.Render("?"),
	)

	return containerStyle.Render(
		joined,
	) */

	return headerStyle.Render(logo)
}
