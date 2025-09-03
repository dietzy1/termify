package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

type device struct {
	Name   string
	ID     string
	Active bool
	Type   string
}

type deviceDisplayModel struct {
	ctx          context.Context
	width        int
	height       int
	spotifyState *state.SpotifyState
	activeDevice *device
}

func newDeviceDisplay(ctx context.Context, spotifyState *state.SpotifyState) deviceDisplayModel {
	return deviceDisplayModel{
		ctx:          ctx,
		width:        28,
		height:       4,
		spotifyState: spotifyState,
		activeDevice: nil,
	}
}

func (m deviceDisplayModel) Init() tea.Cmd {
	return nil
}

func (m *deviceDisplayModel) updateActiveDevice() {
	deviceState := m.spotifyState.GetDeviceState()

	if m.spotifyState == nil || len(deviceState) == 0 {
		m.activeDevice = nil
		return
	}

	for _, d := range deviceState {
		if d.Active {
			m.activeDevice = &device{
				Name:   d.Name,
				ID:     string(d.ID),
				Active: d.Active,
				Type:   d.Type,
			}
			return
		}
	}

	m.activeDevice = nil
}

func (m deviceDisplayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case state.DevicesUpdatedMsg:
		m.updateActiveDevice()
	}

	return m, nil
}

func (m deviceDisplayModel) View() string {
	if m.activeDevice == nil {
		style := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Right).
			Foreground(TextColor)
		return style.Render("No active device")
	}

	nameStyle := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Bold(true).
		Width(m.width).
		PaddingRight(1).
		Foreground(TextColor).
		Italic(true)

	descStyle := lipgloss.NewStyle().
		Foreground(TextColor).
		Align(lipgloss.Right).
		Width(m.width).
		PaddingRight(1)

	deviceInfo := m.activeDevice.Name
	deviceType := m.activeDevice.Type + " • Active"

	return lipgloss.JoinVertical(
		lipgloss.Right,
		nameStyle.Render(deviceInfo),
		descStyle.Render(deviceType),
	)
}
