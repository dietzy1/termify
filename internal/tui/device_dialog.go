package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
	"github.com/zmb3/spotify/v2"
)

type deviceDialogContent struct {
	ctx          context.Context
	width        int
	height       int
	spotifyState *state.SpotifyState
	devices      []device
	cursor       int
}

func NewDeviceDialog(ctx context.Context, spotifyState *state.SpotifyState) DialogContent {
	content := &deviceDialogContent{
		ctx:          ctx,
		spotifyState: spotifyState,
		devices:      []device{},
		cursor:       0,
	}
	content.updateDeviceList()
	return content
}

func (m *deviceDialogContent) Init() tea.Cmd {
	return nil
}

func (m *deviceDialogContent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case state.DevicesUpdatedMsg:
		m.updateDeviceList()
	}
	return m, nil
}

func (m *deviceDialogContent) View() string {
	if len(m.devices) == 0 {
		style := lipgloss.NewStyle().
			Width(40).
			Align(lipgloss.Center).
			Foreground(TextColor)
		return style.Render("No devices found")
	}

	var deviceList []string

	// Base styles for device items
	itemContainerStyle := lipgloss.NewStyle().
		Width(40).
		MarginBottom(1)

	// Name styles
	nameStyle := lipgloss.NewStyle().
		Bold(true).
		Width(40).
		PaddingLeft(2).
		PaddingRight(2).
		Italic(true)

	selectedNameStyle := nameStyle.
		Foreground(WhiteTextColor)

	unselectedNameStyle := nameStyle.
		Foreground(TextColor)

	// Description/type styles
	descStyle := lipgloss.NewStyle().
		Width(40).
		PaddingLeft(2).
		PaddingRight(2).
		Foreground((TextColor))

	for i, device := range m.devices {
		// Prepare device info
		deviceName := device.Name
		deviceType := device.Type
		if device.Active {
			deviceType += " • Active"
		}

		// Add cursor indicator
		var nameText, descText string
		var currentNameStyle lipgloss.Style

		if i == m.cursor {
			nameText = fmt.Sprintf("→ %s", deviceName)
			descText = fmt.Sprintf("  %s", deviceType)
			currentNameStyle = selectedNameStyle
		} else {
			nameText = fmt.Sprintf("  %s", deviceName)
			descText = fmt.Sprintf("  %s", deviceType)
			currentNameStyle = unselectedNameStyle
		}

		// Create two-line device item
		deviceItem := lipgloss.JoinVertical(
			lipgloss.Left,
			currentNameStyle.Render(nameText),
			descStyle.Render(descText),
		)

		deviceList = append(deviceList, itemContainerStyle.Render(deviceItem))
	}

	navHint := lipgloss.NewStyle().
		Foreground(TextColor).
		Italic(true).
		Width(40).
		Align(lipgloss.Center).
		MarginTop(1).
		Render("Use ↑↓ to navigate")

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left, deviceList...),
		navHint,
	)
}

func (m *deviceDialogContent) GetTitle() string {
	return "Select Device"
}

func (m *deviceDialogContent) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *deviceDialogContent) HandleDialogKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.cursor > 0 {
			m.cursor--
		} else if len(m.devices) > 0 {
			m.cursor = len(m.devices) - 1
		}
		return true, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.cursor < len(m.devices)-1 {
			m.cursor++
		} else {
			m.cursor = 0
		}
		return true, nil
	}
	return false, nil
}

func (m *deviceDialogContent) GetActions() []DialogAction {
	selectCmd := func() tea.Msg {
		if len(m.devices) > 0 {
			deviceID := spotify.ID(m.devices[m.cursor].ID)
			// Return a command to select the device
			return m.spotifyState.SelectDevice(m.ctx, deviceID)()
		}
		return DialogMsg{Accepted: true}
	}

	return []DialogAction{
		{
			Label: "Select",
			Key:   key.NewBinding(key.WithKeys("enter")),
			Cmd:   selectCmd,
		},
		{
			Label: "Cancel",
			Key:   key.NewBinding(key.WithKeys("esc")),
			Cmd: func() tea.Msg {
				return DialogMsg{Accepted: false}
			},
		},
	}
}

func (m *deviceDialogContent) updateDeviceList() {
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

	for i, device := range m.devices {
		if device.Active {
			m.cursor = i
			break
		}
	}
}
