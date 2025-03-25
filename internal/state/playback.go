package state

import (
	"context"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

func (s *SpotifyState) FetchPlaybackState() tea.Cmd {
	return func() tea.Msg {
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err: fmt.Errorf("failed to get playback state"),
			}
		}
		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) StartPlayback() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Play(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error starting playback: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)

		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) PausePlayback() tea.Cmd {

	return func() tea.Msg {

		if err := s.client.Pause(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error pausing playback: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)

		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) NextTrack() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Next(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error skipping to next track: %v", err)
			return nil
		}
		//TODO: Implement exponential backoff
		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{
			Err: nil,
		}

	}
}

func (s *SpotifyState) PreviousTrack() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Previous(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error skipping to previous track: %v", err)
			return nil
		}
		//TODO: Implement exponential backoff
		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) PlayTrack(trackID spotify.ID) tea.Cmd {
	return func() tea.Msg {
		if err := s.client.PlayOpt(context.TODO(), &spotify.PlayOptions{
			URIs: []spotify.URI{"spotify:track:" + spotify.URI(trackID)},
		}); err != nil {
			log.Printf("SpotifyState: Error playing track: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) ToggleShuffleMode() tea.Cmd {
	return func() tea.Msg {
		s.mu.RLock()
		shuffleState := s.playerState.ShuffleState
		s.mu.RUnlock()

		if err := s.client.Shuffle(context.TODO(), !shuffleState); err != nil {
			log.Printf("SpotifyState: Error toggling shuffle mode: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) ToggleRepeatMode() tea.Cmd {
	return func() tea.Msg {
		// Determine new repeat state based on current state
		s.mu.RLock()
		currentRepeatState := s.playerState.RepeatState
		s.mu.RUnlock()

		var newState string
		if currentRepeatState == "off" {
			newState = "context"
		} else {
			newState = "off"
		}

		if err := s.client.Repeat(context.TODO(), newState); err != nil {
			log.Printf("SpotifyState: Error setting repeat mode: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.mu.Lock()
		s.playerState = *state
		s.mu.Unlock()

		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}
