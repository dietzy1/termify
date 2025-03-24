package state

import (
	"context"
	"fmt"
	"log"
	"time"

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
		// When we need recommendations, we'll try to base them on the currently playing track
		var seeds spotify.Seeds
		//		limit := 10

		// If we have a currently playing track, use it as a seed
		if s.PlayerState.Item != nil {
			seeds = spotify.Seeds{
				Tracks: []spotify.ID{s.PlayerState.Item.ID},
			}
		} else {
			// Otherwise use default popular tracks as recommendations
			// This is a simplified approach - you might want to use genres or user's top tracks
			log.Println("No current track to base recommendations on")
		}

		// GetRecommendations takes seeds and a map[string]string for attributes
		/* 	attributes := map[string]string{
			"limit": fmt.Sprintf("%d", limit),
		} */

		recommendations, err := s.client.GetRecommendations(context.TODO(), seeds, nil)
		if err != nil {
			log.Printf("SpotifyState: Error fetching recommendations: %v", err)
			return QueueUpdatedMsg{
				Err: fmt.Errorf("failed to get recommendations"),
			}
		}

		log.Printf("SpotifyState: Found %d recommended tracks", len(recommendations.Tracks))

		if len(recommendations.Tracks) == 0 {
			log.Println("SpotifyState: No recommendations returned")
			return QueueUpdatedMsg{
				Err: fmt.Errorf("no recommendations found"),
			}
		}

		// Update Queue with recommendations (keeping as SimpleTrack)
		// We don't need to store the recommendations as FullTrack since we're playing them directly

		// Add recommendations to the queue
		for _, track := range recommendations.Tracks[1:] { // Skip first track as we'll play it directly
			err := s.client.QueueSong(context.TODO(), track.ID)
			if err != nil {
				log.Printf("SpotifyState: Error adding recommendation to queue: %v", err)
			}
		}

		// Play the first recommended track immediately
		firstTrack := recommendations.Tracks[0]
		log.Printf("SpotifyState: Playing first recommended track: %s", firstTrack.Name)

		if err := s.client.PlayOpt(context.TODO(), &spotify.PlayOptions{
			URIs: []spotify.URI{firstTrack.URI},
		}); err != nil {
			log.Printf("SpotifyState: Error playing recommended track: %v", err)
			return QueueUpdatedMsg{
				Err: fmt.Errorf("failed to play recommendation"),
			}
		}

		// Update the queue from Spotify
		queue, err := s.client.GetQueue(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching queue after recommendations: %v", err)
		} else {
			s.Queue = queue.Items
		}

		// Wait for playback to start
		time.Sleep(500 * time.Millisecond)
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
		} else {
			s.PlayerState = *state
		}

		return QueueUpdatedMsg{
			Err: nil,
		}
	}
}
