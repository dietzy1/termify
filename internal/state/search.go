package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

func (s *SpotifyState) SearchEverything(query string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Searching for: %s", query)
		if query == "" {
			log.Printf("SpotifyState: Invalid query")
			return TracksUpdatedMsg{
				Err: fmt.Errorf("invalid query"),
			}
		}

		results, err := s.client.Search(context.TODO(), query, spotify.SearchTypeTrack|spotify.SearchTypeArtist|spotify.SearchTypeAlbum|spotify.SearchTypePlaylist)
		if err != nil {
			log.Printf("SpotifyState: Error searching for tracks: %v", err)
			return TracksUpdatedMsg{
				Err: err,
			}
		}

		log.Printf("SpotifyState: Found %d tracks", len(results.Tracks.Tracks))
		log.Printf("SpotifyState: Found %d artists", len(results.Artists.Artists))
		log.Printf("SpotifyState: Found %d albums", len(results.Albums.Albums))
		log.Printf("SpotifyState: Found %d playlists", len(results.Playlists.Playlists))

		log.Printf("SpotifyState: Found %d tracks", len(s.Tracks))
		return TracksUpdatedMsg{
			Err: nil,
		}
	}
}
