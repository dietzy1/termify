package state

import (
	"context"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

func (s *SpotifyState) FetchPlaybackState(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		state, err := s.client.PlayerState(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return ErrorMsg{
				Title:   "Failed to Fetch Playback State",
				Message: err.Error(),
			}
		}
		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}

func (s *SpotifyState) StartPlayback(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Play(ctx); err != nil {
			log.Printf("SpotifyState: Error starting playback: %v", err)
			return ErrorMsg{
				Title:   "Failed to Start Playback",
				Message: err.Error(),
			}
		}

		time.Sleep(500 * time.Millisecond)

		state, err := s.client.PlayerState(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state after starting: %v", err)
			return ErrorMsg{
				Title:   "Failed to Update Playback State",
				Message: fmt.Sprintf("Playback started, but failed to refresh state: %s", err.Error()),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}

func (s *SpotifyState) PausePlayback(ctx context.Context) tea.Cmd {

	return func() tea.Msg {

		if err := s.client.Pause(ctx); err != nil {
			log.Printf("SpotifyState: Error pausing playback: %v", err)
			return ErrorMsg{
				Title:   "Failed to Pause Playback",
				Message: err.Error(),
			}
		}

		time.Sleep(500 * time.Millisecond)

		state, err := s.client.PlayerState(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state after pausing: %v", err)
			return ErrorMsg{
				Title:   "Failed to Update Playback State",
				Message: fmt.Sprintf("Playback paused, but failed to refresh state: %s", err.Error()),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}

func (s *SpotifyState) NextTrack(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Next(ctx); err != nil {
			log.Printf("SpotifyState: Error skipping to next track: %v", err)
			return ErrorMsg{
				Title:   "Failed to Skip to Next Track",
				Message: err.Error(),
			}
		}

		time.Sleep(500 * time.Millisecond)

		state, err := s.client.PlayerState(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state after skipping next: %v", err)
			return ErrorMsg{
				Title:   "Failed to Update Playback State",
				Message: fmt.Sprintf("Skipped to next track, but failed to refresh state: %s", err.Error()),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}

	}
}

func (s *SpotifyState) PreviousTrack(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Previous(ctx); err != nil {
			log.Printf("SpotifyState: Error skipping to previous track: %v", err)
			return ErrorMsg{
				Title:   "Failed to Skip to Previous Track",
				Message: err.Error(),
			}
		}

		time.Sleep(500 * time.Millisecond)
		state, err := s.client.PlayerState(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state after skipping previous: %v", err)
			return ErrorMsg{
				Title:   "Failed to Update Playback State",
				Message: fmt.Sprintf("Skipped to previous track, but failed to refresh state: %s", err.Error()),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}

func (s *SpotifyState) PlayTrack(ctx context.Context, trackID spotify.ID) tea.Cmd {
	return func() tea.Msg {
		playOptions := &spotify.PlayOptions{
			URIs: []spotify.URI{"spotify:track:" + spotify.URI(trackID)},
		}
		if err := s.client.PlayOpt(ctx, playOptions); err != nil {
			log.Printf("SpotifyState: Error playing track %s: %v", trackID, err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Play Track %s", trackID),
				Message: err.Error(),
			}
		}

		time.Sleep(500 * time.Millisecond)
		state, err := s.client.PlayerState(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state after playing track %s: %v", trackID, err)
			return ErrorMsg{
				Title:   "Failed to Update Playback State",
				Message: fmt.Sprintf("Started playing track %s, but failed to refresh state: %s", trackID, err.Error()),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}

func (s *SpotifyState) ToggleShuffleMode(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		s.mu.RLock()
		shuffleState := s.playerState.ShuffleState
		s.mu.RUnlock()

		newState := !shuffleState
		if err := s.client.Shuffle(ctx, newState); err != nil {
			log.Printf("SpotifyState: Error toggling shuffle mode to %v: %v", newState, err)
			return ErrorMsg{
				Title:   "Failed to Toggle Shuffle",
				Message: err.Error(),
			}
		}

		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state after toggling shuffle: %v", err)
			return ErrorMsg{
				Title:   "Failed to Update Playback State",
				Message: fmt.Sprintf("Toggled shuffle, but failed to refresh state: %s", err.Error()),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}

func (s *SpotifyState) ToggleRepeatMode(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		s.mu.RLock()
		currentRepeatState := s.playerState.RepeatState
		s.mu.RUnlock()

		var newState string
		// Cycle through: off -> context -> track -> off
		switch currentRepeatState {
		case "off":
			newState = "context"
		case "context":
			newState = "track"
		case "track":
			newState = "off"
		default:
			log.Printf("SpotifyState: Unknown repeat state '%s', defaulting to 'off'", currentRepeatState)
			newState = "off"
		}

		if err := s.client.Repeat(ctx, newState); err != nil {
			log.Printf("SpotifyState: Error setting repeat mode to %s: %v", newState, err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Set Repeat Mode to %s", newState),
				Message: err.Error(),
			}
		}

		time.Sleep(500 * time.Millisecond)
		state, err := s.client.PlayerState(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state after setting repeat mode: %v", err)
			return ErrorMsg{
				Title:   "Failed to Update Playback State",
				Message: fmt.Sprintf("Set repeat mode to %s, but failed to refresh state: %s", newState, err.Error()),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{}
	}
}
