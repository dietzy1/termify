package state

import (
	"context"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/zmb3/spotify/v2"
)

type PlaylistsUpdatedMsg struct {
}

type PlaylistSelectedMsg struct {
	PlaylistID string
}

func (s *SpotifyState) FetchPlaylists(ctx context.Context) tea.Cmd {
	log.Printf("SpotifyState: Starting FetchPlaylists command creation")
	return func() tea.Msg {
		log.Printf("SpotifyState: Executing FetchPlaylists, client: %v", s.client != nil)
		if s.client == nil {
			log.Printf("SpotifyState: Error - client is nil")
			return ErrorMsg{
				Title:   "Spotify Client Error",
				Message: "Spotify client not initialized",
			}
		}

		playlists, err := s.client.CurrentUsersPlaylists(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching playlists: %v", err)
			return ErrorMsg{
				Title:   "Failed to Fetch Playlists",
				Message: err.Error(),
			}
		}

		s.mu.Lock()
		s.playlists = playlists.Playlists
		s.mu.Unlock()

		log.Printf("SpotifyState: Successfully fetched %d playlists", len(playlists.Playlists))
		return PlaylistsUpdatedMsg{}
	}
}

func (s *SpotifyState) SelectPlaylist(playlistID string) tea.Cmd {
	return func() tea.Msg {
		// Clean up the playlist ID format
		if strings.HasPrefix(playlistID, "spotify:playlist:") {
			playlistID = strings.Split(playlistID, ":")[2]
		}

		s.mu.Lock()
		s.selectedID = spotify.ID(playlistID)
		s.mu.Unlock()

		log.Printf("SpotifyState: Selected playlist ID: %s", playlistID)
		return PlaylistSelectedMsg{
			PlaylistID: playlistID,
		}
	}
}
