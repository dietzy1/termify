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

//TODO: Since spotify has shitty API support for queueing then it would be more optimal to keep a full client-side queue and just update the queue when the user adds a song to it

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

		s.mu.Lock()
		s.queue = queue.Items
		s.mu.Unlock()

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
