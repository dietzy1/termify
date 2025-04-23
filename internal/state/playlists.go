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
			return ErrorMsg{
				Title:   "Spotify Client Error",
				Message: "Spotify client not initialized",
			}
		}

		playlists, err := s.client.CurrentUsersPlaylists(context.TODO())
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

// FetchPlaylistTracks retrieves all tracks for a given playlist and emits a state update
func (s *SpotifyState) FetchPlaylistTracks(playlistID spotify.ID) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Fetching tracks for playlist: %s", playlistID)
		if playlistID == "" {
			log.Printf("SpotifyState: Invalid playlist ID")
			return ErrorMsg{
				Title:   "Invalid Input",
				Message: "Invalid playlist ID provided",
			}
		}

		s.mu.RLock()
		cachedTracks, exists := s.tracksCache[playlistID]
		s.mu.RUnlock()

		if exists {
			log.Printf("SpotifyState: Found cached tracks for playlist %s", playlistID)
			// Use cached SimpleTracks directly
			s.mu.Lock()
			s.tracks = cachedTracks
			s.mu.Unlock()

			log.Printf("SpotifyState: Successfully loaded %d tracks from cache for playlist %s", len(cachedTracks), playlistID)
			return TracksUpdatedMsg{}
		}

		log.Printf("SpotifyState: No cache found, fetching from API for playlist %s", playlistID)
		playlistItems, err := s.client.GetPlaylistItems(context.TODO(), spotify.ID(playlistID))
		if err != nil {
			log.Printf("SpotifyState: Error fetching playlist items: %v", err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Fetch Tracks for Playlist %s", playlistID),
				Message: err.Error(),
			}
		}

		allTracks := playlistItems.Items

		page := 1
		for {
			log.Printf("SpotifyState: Fetching page %d of playlist items", page)
			err = s.client.NextPage(context.TODO(), playlistItems)
			if err == spotify.ErrNoMorePages {
				break
			}
			if err != nil {
				log.Printf("SpotifyState: Error fetching next page of playlist items: %v", err)
				break
			}

			allTracks = append(allTracks, playlistItems.Items...)
			page++
		}
		log.Printf("SpotifyState: Found %d tracks in total for playlist %s", len(allTracks), playlistID)

		simpleTracks := make([]spotify.SimpleTrack, 0, len(allTracks))
		for _, item := range allTracks {
			if item.Track.Track != nil {
				//Manually map the album name because album for simple track is nil for some reason
				item.Track.Track.SimpleTrack.Album.Name = item.Track.Track.Album.Name
				simpleTracks = append(simpleTracks, item.Track.Track.SimpleTrack)
			}
		}

		s.mu.Lock()
		s.tracks = simpleTracks
		s.tracksCache[playlistID] = simpleTracks
		s.mu.Unlock()

		log.Printf("SpotifyState: Successfully fetched and cached %d tracks for playlist %s", len(simpleTracks), playlistID)

		return TracksUpdatedMsg{}
	}
}

func (s *SpotifyState) SelectPlaylist(playlistID string) tea.Cmd {
	return func() tea.Msg {
		if !strings.HasPrefix(playlistID, "spotify:playlist:") {
			log.Printf("SpotifyState: Invalid playlist ID format: %s", playlistID)
		} else {
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
