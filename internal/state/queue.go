package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type QueueUpdatedMsg struct{}

//TODO: Since spotify has shitty API support for queueing then it would be more optimal to keep a full client-side queue and just update the queue when the user adds a song to it

func (s *SpotifyState) FetchQueue() tea.Cmd {
	return func() tea.Msg {
		queue, err := s.client.GetQueue(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching queue: %v", err)
			return ErrorMsg{
				Title:   "Failed to Fetch Queue",
				Message: err.Error(),
			}
		}
		log.Printf("SpotifyState: Queue has %d items", len(queue.Items))

		s.mu.Lock()
		s.queue = queue.Items
		s.mu.Unlock()

		return QueueUpdatedMsg{}
	}
}

func (s *SpotifyState) AddToQueue(trackId spotify.ID) tea.Cmd {
	return func() tea.Msg {
		if err := s.client.QueueSong(context.TODO(), trackId); err != nil {
			log.Printf("SpotifyState: Error adding track %s to queue: %v", trackId, err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Add Track %s to Queue", trackId),
				Message: err.Error(),
			}
		}

		log.Printf("SpotifyState: Successfully added track %s to queue, fetching updated queue...", trackId)
		return s.FetchQueue()()
	}
}
