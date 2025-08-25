package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
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
	leftSection := m.resizedLogo()
	rightSection := m.navItems()
	availableWidth := min(max(m.width-lipgloss.Width(leftSection)-lipgloss.Width(rightSection), 0), m.width)

	middleSection := lipgloss.NewStyle().
		Width(availableWidth).
		Height(m.height).
		Background(BackgroundColor).Render(" ")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				leftSection,
				middleSection,
				rightSection,
			),
		)
}

func (m navbarModel) navItems() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	helpText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render("? "),
		lipgloss.NewStyle().
			Foreground(TextColor).
			Background(BackgroundColor).
			Render("Help"),
	)

	// New queue text with count
	queueText := lipgloss.JoinHorizontal(lipgloss.Left,
		keyStyle.Render(DefaultKeyMap.ViewQueue.Keys()...),
		lipgloss.NewStyle().
			Foreground(TextColor).
			Background(BackgroundColor).
			PaddingLeft(1).
			Render(fmt.Sprintf("View Queue (%d)", m.queueCount)),
	)

	var paddingTop = 0
	if m.height > 1 {
		paddingTop = 2
	}

	rightSection := lipgloss.JoinHorizontal(lipgloss.Right,
		lipgloss.NewStyle().
			Background(BackgroundColor).
			PaddingTop(paddingTop).
			PaddingRight(2).
			PaddingLeft(2).
			Render(queueText),
		lipgloss.NewStyle().
			Background(BackgroundColor).
			PaddingTop(paddingTop).
			PaddingRight(2).
			PaddingLeft(2).
			Render(helpText),
	)

	return lipgloss.NewStyle().
		Background(BackgroundColor).
		Height(m.height).
		Render(rightSection)
}

func (m navbarModel) resizedLogo() string {
	logoStyle := lipgloss.NewStyle().
		Background(BackgroundColor).
		Foreground(PrimaryColor).
		PaddingLeft(1)

	if m.height > 1 {
		return logoStyle.Render(logo)

	}
	return logoStyle.Render(smallLogo)
}
