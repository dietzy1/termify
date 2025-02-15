package tui

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = (*audioPlayerModel)(nil)

type audioPlayerModel struct {
	width int

	progress int
	//duration int

	bar progress.Model

	spotifyState *SpotifyState
}

func newAudioPlayer(spotifyState *SpotifyState) audioPlayerModel {
	return audioPlayerModel{
		width: 0,
		bar: progress.New(
			progress.WithScaledGradient(PrimaryColor, SecondaryColor),
			progress.WithoutPercentage(),
		),
		spotifyState: spotifyState,
	}
}

func (m audioPlayerModel) songInfoView() string {
	var songTitle string = "Unknown Song"
	if m.spotifyState != nil &&
		m.spotifyState.playerState.Item != nil {
		songTitle = m.spotifyState.playerState.Item.Name
	}

	var artist = "Unknown Artist"
	if m.spotifyState != nil &&
		m.spotifyState.playerState.Item != nil {
		artist = m.spotifyState.playerState.Item.Artists[0].Name
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true).
		Align(lipgloss.Left).
		Width(20).
		Padding(0, 0, 0, 2)

	artistStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		Align(lipgloss.Left).
		Width(20).
		Padding(0, 0, 0, 2)

	songInfo := lipgloss.JoinVertical(
		lipgloss.Top,
		titleStyle.Render(songTitle),
		artistStyle.Render(artist),
	)

	return songInfo
}

func (m audioPlayerModel) View() string {

	volumeSection := lipgloss.NewStyle().
		Width(20).
		Render(m.volumeControlView())

	// Time info styling
	timeInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		Bold(true).
		Align(lipgloss.Center).
		Width(8)

	var duration = 0
	if m.spotifyState != nil &&
		m.spotifyState.playerState.Item != nil {
		duration = int(m.spotifyState.playerState.Item.Duration / 1000)
	} else {
		log.Println("Error getting duration")
	}

	// Create the progress bar
	progressBar := m.bar.ViewAs(float64(m.progress) / float64(duration))

	// Create the time information components
	currentTime := formatDuration(m.progress)
	totalDuration := formatDuration(duration)

	// Create the progress section (times + bar)
	progressSection := lipgloss.JoinHorizontal(
		lipgloss.Left,
		timeInfoStyle.Render(currentTime),
		progressBar,
		timeInfoStyle.Render(totalDuration),
	)

	// Calculate the remaining width for the progress section
	remainingWidth := m.width - lipgloss.Width(m.songInfoView()) - 20 // 20 is the fixed width for volume section

	progressStyle := lipgloss.NewStyle().
		Width(remainingWidth).
		Align(lipgloss.Center)

	// Combine everything horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		m.songInfoView(),
		progressStyle.Render(progressSection),
		volumeSection,
	) + "\n"
}

func (m audioPlayerModel) Init() tea.Cmd {
	return tickCmd()
}

func (m audioPlayerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case PlayerStateUpdatedMsg:
		if msg.Err != nil {
			log.Println("Error updating player state")
			return m, nil
		}
		m.progress = int(msg.State.Progress / 1000)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.bar.Width = msg.Width / 3
		return m, nil

	case tickMsg:
		if m.spotifyState.playerState.Playing {
			m.progress++
			// TODO: We need to handle if the song is paused or stopped

			if m.progress > int(m.spotifyState.playerState.Item.Duration/1000) {
				m.progress = 0
				// Refetch the player state to get the next song
				return m, tea.Batch(
					m.spotifyState.FetchPlaybackState(),
					tickCmd(),
				)
			}

		}

		return m, tickCmd()

	}
	return m, nil
}

func formatDuration(seconds int) string {
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m audioPlayerModel) volumeControlView() string {

	var volume = 80
	// Styles for each component
	iconStyle := lipgloss.NewStyle().
		Width(2).
		Align(lipgloss.Left).MarginRight(1)

	barStyle := lipgloss.NewStyle().
		Width(13).
		Align(lipgloss.Left).
		Foreground(lipgloss.Color(PrimaryColor))

	emptyBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	// Get appropriate volume icon
	var volumeIcon string
	switch {
	case volume == 0:
		volumeIcon = "🔇"
	case volume < 33:
		volumeIcon = "🔈"
	case volume < 66:
		volumeIcon = "🔉"
	default:
		volumeIcon = "🔊"
	}

	// Calculate bar segments
	const (
		volumeChar = "■"
		emptyChar  = "─"
	)
	barWidth := 11 // Fixed width for the bar portion
	filledCount := int(float64(volume) / 100.0 * float64(barWidth))
	emptyCount := barWidth - filledCount

	// Create the volume bar
	filled := strings.Repeat(volumeChar, filledCount)
	empty := emptyBarStyle.Render(strings.Repeat(emptyChar, emptyCount))
	volumeBar := filled + empty

	// Combine all elements using lipgloss
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		iconStyle.Render(volumeIcon),
		barStyle.Render(volumeBar),
	)
}
