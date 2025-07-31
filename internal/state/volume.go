package state

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

// IncreaseVolume increases the volume by 10%
func (s *SpotifyState) IncreaseVolume(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		s.mu.RLock()
		currentVolume := s.playerState.Device.Volume
		s.mu.RUnlock()

		newVolume := min(currentVolume+10, 100)

		err := s.client.Volume(ctx, int(newVolume))
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
func (s *SpotifyState) DecreaseVolume(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		s.mu.RLock()
		currentVolume := s.playerState.Device.Volume
		s.mu.RUnlock()

		// Decrease by 10%, min 0
		newVolume := max(currentVolume-10, 0)

		err := s.client.Volume(ctx, int(newVolume))
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

// Mute toggles the mute state of the player
func (s *SpotifyState) Mute(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		s.mu.RLock()
		currentVolume := s.playerState.Device.Volume
		s.mu.RUnlock()

		var newVolume spotify.Numeric
		if currentVolume > 0 {
			newVolume = 0
		} else {
			newVolume = 50
		}

		err := s.client.Volume(ctx, int(newVolume))
		if err != nil {
			log.Printf("SpotifyState: Error toggling mute: %v", err)
			return ErrorMsg{
				Title:   "Failed to Toggle Mute",
				Message: err.Error(),
			}
		}

		s.mu.Lock()
		s.playerState.Device.Volume = newVolume
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}
