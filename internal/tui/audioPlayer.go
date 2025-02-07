package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/dietzy1/termify/internal/client"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = (*audioPlayerModel)(nil)

type audioPlayerModel struct {
	width int

	status   client.Status
	progress int // Progress in seconds
	bar      progress.Model
}

func newAudioPlayer() audioPlayerModel {

	return audioPlayerModel{
		width: 0,
		bar: progress.New(
			progress.WithScaledGradient(PrimaryColor, SecondaryColor),
			progress.WithoutPercentage(),
		),

		progress: 0,
		status: client.Status{
			Track: client.Track{
				Duration: 200,
			},
		},
	}
}

func (m audioPlayerModel) songInfoView() string {
	// Song Title
	// Artist
	const songTitle = "Shape of you"
	const artist = "Ed Sheeran"

	// Create styling for the song title and artist
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

	// Create the title/artist vertical layout
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

	// Create the progress bar
	progressBar := m.bar.ViewAs(float64(m.progress) / float64(m.status.Track.Duration))

	// Create the time information components
	currentTime := formatDuration(m.progress)
	totalDuration := formatDuration(m.status.Track.Duration)

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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.bar.Width = msg.Width / 3
		return m, nil

	case tickMsg:
		m.progress++
		if m.progress >= m.status.Track.Duration {
			m.progress = 0
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
