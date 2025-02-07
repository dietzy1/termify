package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FocusedModel int

const (
	FocusLibrary FocusedModel = iota
	FocusViewport
	FocusPlaybackControl
)

var focusStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color(PrimaryColor)) /* .
	PaddingRight(1).
	MarginRight(1) */

var unfocusedStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color(BorderColor)) /* .
	PaddingRight(1).
	MarginRight(1) */

const focusModelCount = 3

func (m *model) cycleFocus() {
	m.focusedModel = (m.focusedModel + 1) % focusModelCount
}

func (m *model) cycleFocusBackward() {
	m.focusedModel = (m.focusedModel - 1 + focusModelCount) % focusModelCount
}

func (m model) updateFocusedModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch m.focusedModel {

	case FocusLibrary:
		library, cmd := m.library.Update(msg)
		m.library = library.(libraryModel)
		cmds = append(cmds, cmd)

	case FocusViewport:
		viewport, cmd := m.viewport.Update(msg)
		m.viewport = viewport.(viewportModel)
		cmds = append(cmds, cmd)

	case FocusPlaybackControl:
		playbackControl, cmd := m.playbackControl.Update(msg)
		m.playbackControl = playbackControl.(playbackControlsModel)
		cmds = append(cmds, cmd)

	}
	return m, tea.Batch(cmds...)
}

func applyFocusStyle(isFocused bool) lipgloss.Style {
	if isFocused {
		return focusStyle
	}
	return unfocusedStyle
}
