package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type SearchResultsUpdatedMsg struct {
	Err error
}

func (s *SpotifyState) SearchEverything(query string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Searching for: %s", query)
		if query == "" {
			log.Printf("SpotifyState: Invalid query")
			return SearchResultsUpdatedMsg{
				Err: fmt.Errorf("invalid query"),
			}
		}

		results, err := s.client.Search(context.TODO(), query, spotify.SearchTypeTrack|spotify.SearchTypeArtist|spotify.SearchTypeAlbum|spotify.SearchTypePlaylist)
		if err != nil {
			log.Printf("SpotifyState: Error searching for tracks: %v", err)
			return SearchResultsUpdatedMsg{
				Err: err,
			}
		}

		log.Printf("SpotifyState: Found %d tracks", len(results.Tracks.Tracks))
		log.Printf("SpotifyState: Found %d artists", len(results.Artists.Artists))
		log.Printf("SpotifyState: Found %d albums", len(results.Albums.Albums))
		log.Printf("SpotifyState: Found %d playlists", len(results.Playlists.Playlists))

		// Filter out playlists that are null for some reason
		for i := 0; i < len(results.Playlists.Playlists); i++ {
			if results.Playlists.Playlists[i].Name == "" && results.Playlists.Playlists[i].ID == "" {
				results.Playlists.Playlists = append(results.Playlists.Playlists[:i], results.Playlists.Playlists[i+1:]...)
				i--
			}
		}

		s.mu.Lock()
		s.SearchResults.Tracks = results.Tracks.Tracks
		s.SearchResults.Artists = results.Artists.Artists
		s.SearchResults.Albums = results.Albums.Albums
		s.SearchResults.Playlists = results.Playlists.Playlists

		// Get a local snapshot of the Tracks length for logging
		tracksLength := len(s.Tracks)
		s.mu.Unlock()

		log.Printf("SpotifyState: Found %d tracks", tracksLength)
		return SearchResultsUpdatedMsg{
			Err: nil,
		}
	}
}
