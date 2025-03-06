package state

import (
	"context"
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

func (s *SpotifyState) FetchPlaylists() tea.Cmd {
	log.Printf("SpotifyState: Starting FetchPlaylists command creation")
	return func() tea.Msg {
		log.Printf("SpotifyState: Executing FetchPlaylists, client: %v", s.client != nil)
		if s.client == nil {
			log.Printf("SpotifyState: Error - client is nil")
			return PlaylistsUpdatedMsg{
				Err: fmt.Errorf("spotify client not initialized"),
			}
		}

		playlists, err := s.client.CurrentUsersPlaylists(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playlists: %v", err)
			return PlaylistsUpdatedMsg{
				Err: err,
			}
		}

		s.Playlists = playlists.Playlists
		log.Printf("SpotifyState: Successfully fetched %d playlists", len(s.Playlists))
		return PlaylistsUpdatedMsg{
			Err: nil,
		}
	}
}

// FetchPlaylistTracks retrieves all tracks for a given playlist and emits a state update
func (s *SpotifyState) FetchPlaylistTracks(playlistID string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Fetching tracks for playlist: %s", playlistID)
		if playlistID == "" {
			log.Printf("SpotifyState: Invalid playlist ID")
			return TracksUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		// Check cache first
		if cachedItems, exists := s.playlistItemsCache[playlistID]; exists {
			log.Printf("SpotifyState: Found cached items for playlist %s", playlistID)
			s.Tracks = cachedItems.Items
			return TracksUpdatedMsg{
				Err: nil,
			}
		}

		// If not in cache, fetch from API
		log.Printf("SpotifyState: No cache found, fetching from API for playlist %s", playlistID)
		playlistItems, err := s.client.GetPlaylistItems(context.TODO(), spotify.ID(playlistID))
		if err != nil {
			log.Printf("SpotifyState: Error fetching playlist items: %v", err)
			return TracksUpdatedMsg{
				Err: err,
			}
		}

		// Cache the results
		s.playlistItemsCache[playlistID] = playlistItems
		s.Tracks = playlistItems.Items
		log.Printf("SpotifyState: Successfully fetched and cached %d tracks for playlist %s", len(playlistItems.Items), playlistID)

		return TracksUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) SelectPlaylist(playlistID string) tea.Cmd {
	return func() tea.Msg {
		playlistID = strings.Split(playlistID, ":")[2]
		s.selectedPlaylistID = playlistID
		return PlaylistSelectedMsg{
			PlaylistID: playlistID,
		}
	}
}
