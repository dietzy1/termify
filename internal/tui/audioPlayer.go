package tui

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

var _ tea.Model = (*audioPlayerModel)(nil)

// AutoplayNextTrackMsg is emitted when a song ends and we need to autoplay the next track
type AutoplayNextTrackMsg struct{}

type audioPlayerModel struct {
	ctx   context.Context
	width int

	progress     int
	bar          progress.Model
	spotifyState *state.SpotifyState
}

func newAudioPlayer(ctx context.Context, spotifyState *state.SpotifyState) audioPlayerModel {
	return audioPlayerModel{
		ctx:   ctx,
		width: 0,
		bar: progress.New(
			progress.WithScaledGradient("#1db954", "#212121"),
			progress.WithoutPercentage(),
		),
		spotifyState: spotifyState,
	}
}

func (m audioPlayerModel) songInfoView() string {
	playerState := m.spotifyState.GetPlayerState()

	var songTitle string = "Unknown Song"
	if playerState.Item != nil {
		songTitle = playerState.Item.Name
	}

	var artist = "Unknown Artist"
	if playerState.Item != nil && len(playerState.Item.Artists) > 0 {
		artist = playerState.Item.Artists[0].Name
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Bold(true).
		Align(lipgloss.Left).
		Width(28).
		MaxHeight(5)

	artistStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		Align(lipgloss.Left).
		Width(28)

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
	playerState := m.spotifyState.GetPlayerState()
	if playerState.Item != nil {
		duration = int(playerState.Item.Duration / 1000)
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

	return progressSection
}

func (m audioPlayerModel) Init() tea.Cmd {
	return tickCmd()
}

func (m audioPlayerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case state.PlayerStateUpdatedMsg:
		log.Print("Received PlayerState Update message in audio player layer")
		playerState := m.spotifyState.GetPlayerState()
		m.progress = int(playerState.Progress / 1000)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width

		// Calculate the available width for the progress bar
		songInfoWidth := lipgloss.Width(m.songInfoView())
		volumeControlWidth := lipgloss.Width(m.volumeControlView())
		timeInfoWidth := 8 // 8 for each time display

		// Set the progress bar width based on available space
		m.bar.Width = m.width - songInfoWidth - volumeControlWidth - timeInfoWidth - timeInfoWidth

		if m.width < SHRINKWIDTH {
			m.bar.Width = m.width - songInfoWidth - volumeControlWidth - timeInfoWidth - timeInfoWidth
		}

		return m, nil

	case tickMsg:
		playerState := m.spotifyState.GetPlayerState()
		if playerState.Playing {
			m.progress++

			if m.progress > int(playerState.Item.Duration/1000) {
				m.progress = 0
				return m, tea.Batch(
					func() tea.Msg { return AutoplayNextTrackMsg{} },
					tickCmd(),
				)
			}
		}
		return m, tickCmd()

	}
	return m, nil
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
		playerState := m.spotifyState.GetPlayerState()
		volume = int(playerState.Device.Volume)
	}

	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor)).
		PaddingRight(2)

	emptyBarStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor))

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

	// Create a container with fixed width and left alignment
	containerStyle := lipgloss.NewStyle().
		Width(28).
		Align(lipgloss.Right).
		MaxHeight(1)

	if m.width < SHRINKWIDTH {
		containerStyle = containerStyle.Width(20)
	}

	return containerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			volumeIcon,
			barStyle.Render(volumeBar),
		),
	)
}
