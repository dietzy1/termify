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

	progress     int
	bar          progress.Model
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
		Width(28).
		MarginLeft(2)

	artistStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		Align(lipgloss.Left).
		Width(28).
		MarginLeft(2)

	if m.width < SHRINKWIDTH {
		titleStyle = titleStyle.MarginLeft(0)
		artistStyle = artistStyle.MarginLeft(0)
		titleStyle = titleStyle.Width(20)
		artistStyle = artistStyle.Width(20)
	}

	songInfo := lipgloss.JoinVertical(
		lipgloss.Top,
		titleStyle.Render(songTitle),
		artistStyle.Render(artist),
	)

	return songInfo
}

func (m audioPlayerModel) View() string {
	timeInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		Bold(true).
		Align(lipgloss.Center).
		Width(8)

	var duration = 0
	if m.spotifyState != nil &&
		m.spotifyState.playerState.Item != nil {
		duration = int(m.spotifyState.playerState.Item.Duration / 1000)
	}

	// Create the progress bar
	progressBar := m.bar.ViewAs(float64(m.progress) / float64(duration))

	// Create the time information components
	currentTime := formatDuration(m.progress)
	totalDuration := formatDuration(duration)

	// Create the progress section (times + bar)
	progressSection := lipgloss.JoinHorizontal(
		lipgloss.Center,
		timeInfoStyle.Render(currentTime),
		progressBar,
		timeInfoStyle.Render(totalDuration),
	)

	remainingWidth := m.width - lipgloss.Width(m.songInfoView()) - lipgloss.Width(m.volumeControlView())

	progressStyle := lipgloss.NewStyle().
		Width(remainingWidth).
		Align(lipgloss.Center)
	return progressStyle.Render(progressSection)
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
		if m.width < SHRINKWIDTH {
			m.bar.Width = msg.Width - lipgloss.Width(m.songInfoView()) - lipgloss.Width(m.volumeControlView()) - 20
		}
		return m, nil

	case tickMsg:
		if m.spotifyState.playerState.Playing {
			m.progress++

			if m.progress > int(m.spotifyState.playerState.Item.Duration/1000) {
				m.progress = 0
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
	var volume = 0
	if m.spotifyState != nil {
		volume = int(m.spotifyState.playerState.Device.Volume)
	}

	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		MarginRight(2)

	emptyBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

	if m.width > SHRINKWIDTH {
		barStyle.UnsetMarginRight()
	}

	var volumeIcon string
	switch {
	case volume == 0:
		volumeIcon = "🔇 "
	case volume < 33:
		volumeIcon = "🔈 "
	case volume < 66:
		volumeIcon = "🔉 "
	default:
		volumeIcon = "🔊 "
	}

	const (
		volumeChar = "■"
		emptyChar  = "─"
	)

	barWidth := 10
	filledCount := volume / 10
	emptyCount := barWidth - filledCount

	// Create the volume bar
	filled := strings.Repeat(volumeChar, filledCount)
	empty := emptyBarStyle.Render(strings.Repeat(emptyChar, emptyCount))
	volumeBar := filled + empty

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		volumeIcon,
		barStyle.Render(volumeBar),
	)
}
