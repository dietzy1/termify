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

	helpText := "? Help "
	//settings := "⚙ Settings "

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		Align(lipgloss.Center).
		Width(m.width - 7 - lipgloss.Width(helpText)).
		Background(lipgloss.Color(BackgroundColor))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		Background(lipgloss.Color(BackgroundColor)).
		Bold(true).
		Height(lipgloss.Height(logo)).
		Align(lipgloss.Center, lipgloss.Center)

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Bottom, helpStyle.Render("       "), headerStyle.Render(logo), helpStyle.Render(helpText)),
	)
}
