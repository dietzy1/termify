package state

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

// IncreaseVolume increases the volume by 10%
func (s *SpotifyState) IncreaseVolume() tea.Cmd {
	return func() tea.Msg {
		s.mu.RLock()
		currentVolume := s.playerState.Device.Volume
		s.mu.RUnlock()

		// Increase by 10%, max 100
		newVolume := currentVolume + 10
		if newVolume > 100 {
			newVolume = 100
		}

		err := s.client.Volume(context.TODO(), int(newVolume))
		if err != nil {
			log.Printf("SpotifyState: Error increasing volume: %v", err)
			return ErrorMsg{
				Title:   "Failed to Increase Volume",
				Message: err.Error(),
			}
		}

		s.mu.Lock()
		s.playerState.Device.Volume = newVolume
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}

// DecreaseVolume decreases the volume by 10%
func (s *SpotifyState) DecreaseVolume() tea.Cmd {
	return func() tea.Msg {
		s.mu.RLock()
		currentVolume := s.playerState.Device.Volume
		s.mu.RUnlock()

		// Decrease by 10%, min 0
		newVolume := currentVolume - 10
		if newVolume < 0 {
			newVolume = 0
		}

		err := s.client.Volume(context.TODO(), int(newVolume))
		if err != nil {
			log.Printf("SpotifyState: Error decreasing volume: %v", err)
			return ErrorMsg{
				Title:   "Failed to Decrease Volume",
				Message: err.Error(),
			}
		}

		s.mu.Lock()
		s.playerState.Device.Volume = newVolume
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}
