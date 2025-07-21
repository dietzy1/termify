package state

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

const (
	maxRetries     = 3
	baseRetryDelay = 200 * time.Millisecond
	maxRetryDelay  = 2 * time.Second
)

type RetryableOperation func(ctx context.Context) error

func (s *SpotifyState) executeWithStateUpdate(ctx context.Context, operation RetryableOperation, operationName string) tea.Cmd {
	return func() tea.Msg {

		if err := operation(ctx); err != nil {
			log.Printf("SpotifyState: Error in %s: %v", operationName, err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to %s", operationName),
				Message: err.Error(),
			}
		}

		if err := s.updateStateWithRetry(ctx); err != nil {
			log.Printf("SpotifyState: Error updating state after %s: %v", operationName, err)
			return ErrorMsg{
				Title:   "Failed to Update Playback State",
				Message: fmt.Sprintf("%s completed, but failed to refresh state: %s", operationName, err.Error()),
			}
		}

		return PlayerStateUpdatedMsg{}
	}
}

// updateStateWithRetry fetches player state with exponential backoff for unchanged timestamps
func (s *SpotifyState) updateStateWithRetry(ctx context.Context) error {

	s.mu.RLock()
	lastTimestamp := s.playerState.Timestamp
	s.mu.RUnlock()

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		state, err := s.client.PlayerState(ctx)
		if err != nil {
			lastErr = err
			log.Printf("SpotifyState: API error on attempt %d: %v", attempt+1, err)

			delay := time.Duration(float64(baseRetryDelay) * math.Pow(2, float64(attempt)))
			if delay > maxRetryDelay {
				delay = maxRetryDelay
			}
			time.Sleep(delay)
			continue
		}

		log.Printf("SpotifyState: Player state fetched (attempt %d): timestamp %d vs %d",
			attempt+1, state.Timestamp, lastTimestamp)

		if state.Timestamp != lastTimestamp {
			s.mu.Lock()
			s.playerState = *state
			s.mu.Unlock()
			return nil
		}

		lastErr = fmt.Errorf("player state timestamp unchanged after %d attempts", attempt+1)

		if attempt < maxRetries-1 {
			delay := time.Duration(float64(baseRetryDelay) * math.Pow(2, float64(attempt)))
			if delay > maxRetryDelay {
				delay = maxRetryDelay
			}
			log.Printf("SpotifyState: Timestamp unchanged on attempt %d, waiting %v before retry", attempt+1, delay)
			time.Sleep(delay)
		}
	}

	log.Printf("SpotifyState: Max retries reached, using last fetched state")
	return lastErr
}

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
	operation := func(ctx context.Context) error {
		return s.client.Play(ctx)
	}
	return s.executeWithStateUpdate(ctx, operation, "Start Playback")
}

func (s *SpotifyState) PausePlayback(ctx context.Context) tea.Cmd {
	operation := func(ctx context.Context) error {
		return s.client.Pause(ctx)
	}
	return s.executeWithStateUpdate(ctx, operation, "Pause Playback")
}

func (s *SpotifyState) NextTrack(ctx context.Context) tea.Cmd {
	operation := func(ctx context.Context) error {
		return s.client.Next(ctx)
	}
	return s.executeWithStateUpdate(ctx, operation, "Skip to Next Track")
}

func (s *SpotifyState) PreviousTrack(ctx context.Context) tea.Cmd {
	operation := func(ctx context.Context) error {
		return s.client.Previous(ctx)
	}
	return s.executeWithStateUpdate(ctx, operation, "Skip to Previous Track")
}

func (s *SpotifyState) PlayTrack(ctx context.Context, trackID spotify.ID) tea.Cmd {
	operation := func(ctx context.Context) error {
		playOptions := &spotify.PlayOptions{
			URIs: []spotify.URI{"spotify:track:" + spotify.URI(trackID)},
		}
		return s.client.PlayOpt(ctx, playOptions)
	}
	return s.executeWithStateUpdate(ctx, operation, fmt.Sprintf("Play Track %s", trackID))
}

func (s *SpotifyState) ToggleShuffleMode(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		s.mu.RLock()
		shuffleState := s.playerState.ShuffleState
		s.mu.RUnlock()

		newState := !shuffleState
		operation := func(ctx context.Context) error {
			return s.client.Shuffle(ctx, newState)
		}

		cmd := s.executeWithStateUpdate(ctx, operation, "Toggle Shuffle")
		return cmd()
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

		operation := func(ctx context.Context) error {
			return s.client.Repeat(ctx, newState)
		}

		cmd := s.executeWithStateUpdate(ctx, operation, fmt.Sprintf("Set Repeat Mode to %s", newState))
		return cmd()
	}
}
