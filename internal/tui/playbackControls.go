package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = (*playbackControlsModel)(nil)

type playbackControlsModel struct {
	width         int
	currentButton int
	volume        float64

	spotifyState *SpotifyState
}

func newPlaybackControlsModel(spotifyState *SpotifyState) playbackControlsModel {
	return playbackControlsModel{
		currentButton: 1,
		volume:        0.5,
		width:         0,
		spotifyState:  spotifyState,
	}
}

func (m playbackControlsModel) Init() tea.Cmd {
	return nil
}

func (m playbackControlsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case PlayerStateUpdatedMsg:
		log.Println("Playback controls recieved player state update")
		if msg.Err != nil {
			log.Println("Error updating player state")
			return m, nil
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Left):
			m.currentButton = (m.currentButton - 1 + 3) % 3
		case key.Matches(msg, DefaultKeyMap.Right):
			m.currentButton = (m.currentButton + 1) % 3
		case key.Matches(msg, DefaultKeyMap.Select):
			if m.currentButton == 0 {
				return m, m.spotifyState.PreviousTrack()
			}
			// Play/Pause button
			if m.currentButton == 1 {
				if m.spotifyState.playerState.Playing {
					return m, m.spotifyState.PausePlayback()
				}
				if !m.spotifyState.playerState.Playing {
					return m, m.spotifyState.StartPlayback()
				}
			}
			//Skip next button
			if m.currentButton == 2 {
				return m, m.spotifyState.NextTrack()
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

	if m.spotifyState.playerState.Playing {
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
