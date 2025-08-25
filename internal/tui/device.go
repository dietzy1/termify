package tui

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/dietzy1/termify/internal/state"
	"github.com/zmb3/spotify/v2"
)

type device struct {
	Name   string
	ID     string
	Active bool
	Type   string
}

type deviceKeyMap struct {
	Previous key.Binding
	Next     key.Binding
	Select   key.Binding
	Escape   key.Binding
}

type deviceSelectorModel struct {
	ctx          context.Context
	width        int
	height       int
	spotifyState *state.SpotifyState
	devices      []device
	cursor       int
	isFocused    bool
}

func NewDeviceSelector(ctx context.Context, spotifyState *state.SpotifyState) deviceSelectorModel {
	return deviceSelectorModel{
		ctx:          ctx,
		height:       4,
		width:        29,
		spotifyState: spotifyState,
		devices:      []device{},
		cursor:       0,
		isFocused:    false,
	}
}

func (m deviceSelectorModel) Init() tea.Cmd {

	if len(m.devices) == 0 {
		log.Println("No devices")
		return nil
	}

	deviceID := spotify.ID(m.devices[m.cursor].ID)
	if deviceID == "" {
		log.Println("Unable to retrieve device ID")
		return nil
	}
	return m.spotifyState.SelectDevice(m.ctx, deviceID)
}

func (m *deviceSelectorModel) Focus() {
	m.isFocused = true
	// Find and select the active device
	for i, device := range m.devices {
		if device.Active {
			m.cursor = i
			break
		}
	}
}

func (m *deviceSelectorModel) Blur() {
	m.isFocused = false
}

func (m *deviceSelectorModel) updateDeviceList() {
	deviceState := m.spotifyState.GetDeviceState()

	if m.spotifyState == nil || len(deviceState) == 0 {
		m.devices = []device{}
		return
	}

	m.devices = make([]device, 0, len(deviceState))
	for _, d := range deviceState {
		m.devices = append(m.devices, device{
			Name:   d.Name,
			ID:     string(d.ID),
			Active: d.Active,
			Type:   d.Type,
		})
	}

	// Find and select the active device
	for i, device := range m.devices {
		if device.Active {
			m.cursor = i
			break
		}
	}
}

func (m deviceSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case state.DevicesUpdatedMsg:
		m.updateDeviceList()

	case tea.KeyMsg:
		if !m.isFocused {
			return m, nil
		}

		switch {
		case key.Matches(msg, DefaultKeyMap.Left):
			if m.cursor > 0 {
				m.cursor--
			} else if len(m.devices) > 0 {
				// Wrap around to the end
				m.cursor = len(m.devices) - 1
			}
		case key.Matches(msg, DefaultKeyMap.Right):
			if m.cursor < len(m.devices)-1 {
				m.cursor++
			} else {
				// Wrap around to the beginning
				m.cursor = 0
			}
		case key.Matches(msg, DefaultKeyMap.Select):
			if len(m.devices) == 0 {
				return m, nil
			}
			deviceID := spotify.ID(m.devices[m.cursor].ID)
			return m, m.spotifyState.SelectDevice(m.ctx, deviceID)
		}
	}

	return m, nil
}

func (m deviceSelectorModel) View() string {
	if len(m.devices) == 0 {
		style := lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Right).
			Background(BackgroundColor)
		return style.Render("No devices found")
	}

	currentDevice := m.devices[m.cursor]

	navStyle := lipgloss.NewStyle().
		Foreground(TextColor).
		Align(lipgloss.Right).
		PaddingRight(2).
		Background(BackgroundColor).
		Width(30)

	navText := ""
	if len(m.devices) > 1 {
		navText = fmt.Sprintf("%d/%d", m.cursor+1, len(m.devices))
	}

	nameStyle := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Bold(true).
		Width(m.width).
		PaddingRight(1).
		BorderRight(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderBackground(BackgroundColor).
		BorderForeground(getBorderStyle(m.isFocused)).
		Background(BackgroundColor)

	descStyle := lipgloss.NewStyle().
		Foreground(TextColor).
		Align(lipgloss.Right).
		Width(m.width).
		PaddingRight(1).
		BorderRight(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderBackground(BackgroundColor).
		BorderForeground(getBorderStyle(m.isFocused)).
		Background(BackgroundColor)

	deviceInfo := currentDevice.Name
	deviceType := currentDevice.Type
	if currentDevice.Active {
		deviceType += " • Active"
		nameStyle = nameStyle.Foreground(TextColor).Italic(true)
	}
	if !currentDevice.Active {
		deviceType += " • Inactive"
	}
	if m.isFocused {
		nameStyle = nameStyle.Foreground(PrimaryColor)
	}

	joined := lipgloss.JoinVertical(
		lipgloss.Right,
		nameStyle.Render(deviceInfo),
		descStyle.Render(deviceType),
		navStyle.Render(navText),
	)

	// Use lipgloss.Place to get background colored
	return lipgloss.Place(30, 4,
		lipgloss.Right,
		lipgloss.Bottom,
		joined,
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(BackgroundColor)),
	)
}
