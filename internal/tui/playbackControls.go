package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zmb3/spotify/v2"
)

var _ tea.Model = (*playbackControlsModel)(nil)

type playbackControlsModel struct {
	width         int
	currentButton int
	//isPlaying     bool
	volume float64

	spotifyState *SpotifyState
}

func newPlaybackControlsModel(spotifyState *SpotifyState) playbackControlsModel {
	return playbackControlsModel{
		currentButton: 1,
		//isPlaying:     false,
		volume:       0.5,
		width:        0,
		spotifyState: spotifyState,
	}
}

func (m playbackControlsModel) Init() tea.Cmd {
	return nil
}

func (m playbackControlsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StateUpdateMsg:
		if msg.Type == PlayerStateUpdated {
			log.Println("Playback controls recieved player state update")
			if msg.Err != nil {
				log.Println("Error updating player state")
				return m, nil
			}
			if playerState, ok := msg.Data.(*spotify.PlayerState); ok {
				log.Println("Updating playback controls with new player state", playerState)
				m.spotifyState.playerState.Playing = playerState.Playing
			} else {
				log.Println("Error converting data to player state")
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			m.currentButton = (m.currentButton - 1 + 3) % 3
		case "right":
			m.currentButton = (m.currentButton + 1) % 3
		case "enter", " ":
			//Skip previous button
			if m.currentButton == 0 {
				return m, m.spotifyState.PreviousTrack()
			}

			// Play/Pause button
			if m.currentButton == 1 {
				m.spotifyState.playerState.Playing = !m.spotifyState.playerState.Playing
				/* return m, PlayPause() */
				if m.spotifyState.playerState.Playing {
					return m, m.spotifyState.StartPlayback()
				}
				if !m.spotifyState.playerState.Playing {
					return m, m.spotifyState.PausePlayback()
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
