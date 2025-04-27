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

		// Remove consecutive duplicate tracks if they appear 10 or more times in a row
		if len(queue.Items) > 0 {
			var deduped []spotify.FullTrack
			current := queue.Items[0]
			count := 1

			for i := 1; i < len(queue.Items); i++ {
				if queue.Items[i].ID == current.ID {
					count++
				} else {
					if count < 5 {
						deduped = append(deduped, current)
					}
					current = queue.Items[i]
					count = 1
				}
			}
			// Add the last track if it doesn't have too many duplicates
			if count < 10 {
				deduped = append(deduped, current)
			}
			queue.Items = deduped
		}

		s.mu.Lock()
		s.queue = queue.Items
		s.mu.Unlock()

		return QueueUpdatedMsg{}
	}
}

func (s *SpotifyState) AddToQueue(trackId spotify.ID) tea.Cmd {

	if err := s.client.QueueSong(context.TODO(), trackId); err != nil {
		log.Printf("SpotifyState: Error adding track %s to queue: %v", trackId, err)
		return func() tea.Msg {
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Add Track %s to Queue", trackId),
				Message: err.Error(),
			}
		}
	}

	log.Printf("SpotifyState: Successfully added track %s to queue, fetching updated queue...", trackId)
	return s.FetchQueue()
}
