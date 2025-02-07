package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = (*playbackControlsModel)(nil)

type playbackControlsModel struct {
	width         int
	currentButton int
	isPlaying     bool
	volume        float64
}

func newPlaybackControlsModel() playbackControlsModel {
	return playbackControlsModel{
		currentButton: 1,
		isPlaying:     false,
		volume:        0.5,
		width:         0,
	}
}

func (m playbackControlsModel) Init() tea.Cmd {
	return nil
}

func (m playbackControlsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			m.currentButton = (m.currentButton - 1 + 3) % 3
		case "right":
			m.currentButton = (m.currentButton + 1) % 3
		case "enter", " ":
			if m.currentButton == 1 {
				m.isPlaying = !m.isPlaying
			}
		}

	}
	return m, nil
}

func (m playbackControlsModel) View() string {

	var baseStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		Bold(true).
		Padding(0, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(BorderColor))

	var activeStyle = baseStyle.
		Foreground(lipgloss.Color(PrimaryColor)).
		BorderForeground(lipgloss.Color(PrimaryColor))

	prevButton := "⏮"
	playPauseButton := "▶"
	nextButton := "⏭"

	if m.isPlaying {
		playPauseButton = "⏸"
	}

	buttons := []string{prevButton, playPauseButton, nextButton}
	var renderedButtons []string

	for i, button := range buttons {
		if i == m.currentButton {
			renderedButtons = append(renderedButtons, activeStyle.Render(button))
		} else {
			renderedButtons = append(renderedButtons, baseStyle.Render(button))
		}
	}

	playbackControls := lipgloss.JoinHorizontal(lipgloss.Bottom, renderedButtons...)

	containerStyle := lipgloss.NewStyle().
		Width(m.width).
		Foreground(lipgloss.Color(TextColor)).
		Align(lipgloss.Center, lipgloss.Center)

	return containerStyle.Render(
		playbackControls,
	)
}
