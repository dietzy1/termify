package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FocusedModel int

const (
	FocusLibrary FocusedModel = iota
	FocusSearchBar
	FocusViewport
)

const focusModelCount = 3

func (m *applicationModel) cycleFocus() {
	m.focusedModel = (m.focusedModel + 1) % focusModelCount
}

func (m *applicationModel) cycleFocusBackward() {
	m.focusedModel = (m.focusedModel - 1 + focusModelCount) % focusModelCount
}

func (m applicationModel) updateFocusedModel(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case FocusSearchBar:
		searchBar, cmd := m.searchBar.Update(msg)
		m.searchBar = searchBar.(searchbarModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// Helper function to get border style based on focus state
func getBorderStyle(isFocused bool) lipgloss.Color {
	if isFocused {
		return lipgloss.Color(PrimaryColor)
	}
	return lipgloss.Color(BorderColor)
}
