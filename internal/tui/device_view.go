package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

// Device implements list.Item interface for Spotify devices
type Device struct {
	Name       string
	ID         string
	Active     bool
	Type       string
	Volume     int
	Restricted bool
}

// Title returns just the device name
func (d Device) Title() string {
	return d.Name
}

// Description returns device type and active status
func (d Device) Description() string {
	if d.Active {
		return " " + d.Type + " • Active"
	}
	return d.Type
}

// FilterValue returns the string to filter on
func (d Device) FilterValue() string { return d.Name }

// DeviceSelectorModel is the main component
type DeviceSelectorModel struct {
	spotifyState *state.SpotifyState
	list         list.Model
}

// NewDeviceSelector creates a new device selector component
func NewDeviceSelector(spotifyState *state.SpotifyState) DeviceSelectorModel {
	// Create delegate with custom styles
	delegate := list.NewDefaultDelegate()
	const itemWidth = 28

	// Style the list elements

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		/* Padding(0, 0, 0, 2). */
		Width(itemWidth)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, false). // All borders disabled
		BorderForeground(lipgloss.Color(PrimaryColor)).
		Foreground(lipgloss.Color(PrimaryColor)).
		Bold(true).
		Padding(0, 0, 0, 1).
		Width(itemWidth)

	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		/* Padding(0, 0, 0, 2). */
		Width(itemWidth)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color(TextColor)).
		/* Padding(0, 0, 0, 2). */
		Width(itemWidth)

	// Create the list
	l := list.New([]list.Item{}, delegate, itemWidth+2, 1) // only show 1 item at a time
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(true)
	l.DisableQuitKeybindings()

	model := DeviceSelectorModel{
		spotifyState: spotifyState,
		list:         l,
	}

	return model
}

// Init initializes the model
func (m DeviceSelectorModel) Init() tea.Cmd {
	return nil
}

// updateDeviceList converts the spotify devices to list items
/* func (m *DeviceSelectorModel) updateDeviceList() {
	if m.spotifyState == nil || len(m.spotifyState.DeviceState) == 0 {
		return
	}

	items := make([]list.Item, 0, len(m.spotifyState.DeviceState))
	for _, d := range m.spotifyState.DeviceState {
		items = append(items, Device{
			Name:       d.Name,
			ID:         string(d.ID),
			Active:     d.Active,
			Type:       d.Type,
			Volume:     int(d.Volume),
			Restricted: d.Restricted,
		})
	}
	for _, d := range m.spotifyState.DeviceState {
		items = append(items, Device{
			Name:       d.Name,
			ID:         string(d.ID),
			Active:     d.Active,
			Type:       d.Type,
			Volume:     int(d.Volume),
			Restricted: d.Restricted,
		})
	}
	m.list.SetItems(items)
} */

// Update handles messages and user input
func (m DeviceSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always update device list from the current state
	//	m.updateDeviceList()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Select):
			// Get the selected device and transfer playback
			if len(m.list.Items()) > 0 {
				selectedDevice := m.list.SelectedItem().(Device)
				if !selectedDevice.Restricted {
					// Here you would implement transferring playback to the device
					return m, tea.Println("Transferring playback to: " + selectedDevice.Name)
				} else {
					return m, tea.Println("Cannot transfer to restricted device: " + selectedDevice.Name)
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI
func (m DeviceSelectorModel) View() string {
	// Update device list before rendering
	//m.updateDeviceList()

	if len(m.list.Items()) == 0 {
		return "No devices found."
	}
	return m.list.View()
}

func (m applicationModel) viewDevice() string {
	playbackSection := m.renderPlaybackSection()

	return lipgloss.JoinVertical(
		lipgloss.Center,
		m.navbar.View(),
		lipgloss.JoinVertical(lipgloss.Top,
			m.deviceSelector.View(),
			playbackSection,
			"\r",
		))
}
