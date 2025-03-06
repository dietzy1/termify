package state

import (
	"context"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

		s.PlayerState = *state
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

		s.PlayerState = *state

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

		s.PlayerState = *state

		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}
