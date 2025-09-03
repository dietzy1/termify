package tui

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

// Helper function to update submodels
func updateSubmodel[T any](model tea.Model, msg tea.Msg, targetType T) (T, tea.Cmd, bool) {
	updatedModel, cmd := model.Update(msg)

	typedModel, ok := updatedModel.(T)
	if !ok {
		log.Printf("Error: failed to convert %T to %T\n", updatedModel, targetType)
		return targetType, tea.Quit, false
	}

	return typedModel, cmd, true
}

func updateAndAssign[T tea.Model](target *T, msg tea.Msg) tea.Cmd {
	updated, cmd := (*target).Update(msg)
	*target = updated.(T)
	return cmd
}

func formatDuration(seconds int) string {
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}

// Helper function to format track duration
func formatTrackDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}

func safelyRenderError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
