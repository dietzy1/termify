package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type QueueUpdatedMsg struct {
	Err error
}

func (s *SpotifyState) FetchQueue() tea.Cmd {
	return func() tea.Msg {
		queue, err := s.client.GetQueue(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching queue: %v", err)
			return QueueUpdatedMsg{
				Err: fmt.Errorf("failed to get queue"),
			}
		}
		log.Println("SpotifyState: Queue:", queue)

		s.Queue = queue.Items
		return QueueUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) AddToQueue(trackId spotify.ID) tea.Cmd {
	return func() tea.Msg {
		if err := s.client.QueueSong(context.TODO(), trackId); err != nil {
			log.Printf("SpotifyState: Error adding to queue: %v", err)
			return QueueUpdatedMsg{
				Err: fmt.Errorf("failed to add to queue"),
			}
		}

		return s.FetchQueue()()
	}
}

func (s *SpotifyState) FetchRecommendations() tea.Cmd {
	return func() tea.Msg {
		recommendations, err := s.client.GetRecommendations(context.TODO(), spotify.Seeds{}, nil)
		if err != nil {
			log.Printf("SpotifyState: Error fetching recommendations: %v", err)
			return QueueUpdatedMsg{
				Err: fmt.Errorf("failed to get recommendations"),
			}
		}

		log.Println("SpotifyState: Recommendations:", recommendations.Tracks)
		// Add some stuff to queue

		return QueueUpdatedMsg{
			Err: nil,
		}
	}
}
