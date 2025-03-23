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
func (s *SpotifyState) FetchPlaylistTracks(playlistID spotify.ID) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Fetching tracks for playlist: %s", playlistID)
		if playlistID == "" {
			log.Printf("SpotifyState: Invalid playlist ID")
			return TracksUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		if cachedTracks, exists := s.tracksCache[playlistID]; exists {
			log.Printf("SpotifyState: Found cached tracks for playlist %s", playlistID)
			// Use cached SimpleTracks directly
			s.Tracks = cachedTracks
			log.Printf("SpotifyState: Successfully loaded %d tracks from cache for playlist %s", len(cachedTracks), playlistID)
			return TracksUpdatedMsg{
				Err: nil,
			}
		}

		log.Printf("SpotifyState: No cache found, fetching from API for playlist %s", playlistID)
		playlistItems, err := s.client.GetPlaylistItems(context.TODO(), spotify.ID(playlistID))
		if err != nil {
			log.Printf("SpotifyState: Error fetching playlist items: %v", err)
			return TracksUpdatedMsg{
				Err: err,
			}
		}

		allTracks := playlistItems.Items

		for page := 1; ; page++ {
			log.Printf("SpotifyState: Fetching page %d of playlist items", page)
			err = s.client.NextPage(context.TODO(), playlistItems)
			if err == spotify.ErrNoMorePages {
				break
			}
			if err != nil {
				log.Println("SpotifyState: Error fetching next page of playlist items:", err)
				break
			}

			allTracks = append(allTracks, playlistItems.Items...)
		}
		log.Println("SpotifyState: Found", len(allTracks), "tracks")

		simpleTracks := make([]spotify.SimpleTrack, 0, len(allTracks))
		for _, item := range allTracks {
			if item.Track.Track != nil {
				//Manually map the album name because album for simple track is nil for some reason
				item.Track.Track.SimpleTrack.Album.Name = item.Track.Track.Album.Name
				simpleTracks = append(simpleTracks, item.Track.Track.SimpleTrack)
			}
		}

		s.Tracks = simpleTracks
		s.tracksCache[playlistID] = simpleTracks
		log.Printf("SpotifyState: Successfully fetched and cached %d tracks for playlist %s", len(simpleTracks), playlistID)

		return TracksUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) SelectPlaylist(playlistID string) tea.Cmd {
	return func() tea.Msg {
		playlistID = strings.Split(playlistID, ":")[2]
		s.SelectedID = spotify.ID(playlistID)
		return PlaylistSelectedMsg{
			PlaylistID: playlistID,
		}
	}
}
