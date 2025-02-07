package tui

import (
	"context"
	"log"

	"github.com/zmb3/spotify/v2"
)

// SpotifyState manages all Spotify-related state and API calls
type SpotifyState struct {
	client *spotify.Client

	// Cache for current data
	playlists []spotify.SimplePlaylist
	tracks    []spotify.PlaylistTrack

	// Currently selected items
	selectedPlaylistID string
}

func NewSpotifyState(client *spotify.Client) *SpotifyState {
	return &SpotifyState{
		client: client,
	}
}

// FetchPlaylists retrieves all playlists for the current user
func (s *SpotifyState) FetchPlaylists() ([]spotify.SimplePlaylist, error) {
	playlists, err := s.client.CurrentUsersPlaylists(context.Background())
	if err != nil {
		return nil, err
	}
	s.playlists = playlists.Playlists
	return s.playlists, nil
}

// FetchPlaylistTracks retrieves all tracks for a given playlist
func (s *SpotifyState) FetchPlaylistTracks(playlistID string) ([]spotify.PlaylistTrack, error) {
	if playlistID == "" {
		return nil, nil
	}

	tracks, err := s.client.GetPlaylistTracks(context.Background(), spotify.ID(playlistID))
	if err != nil {
		log.Printf("Error fetching tracks for playlist %s: %v", playlistID, err)
		return nil, err
	}

	s.tracks = tracks.Tracks
	s.selectedPlaylistID = playlistID
	return s.tracks, nil
}

// GetCurrentTracks returns the currently loaded tracks
func (s *SpotifyState) GetCurrentTracks() []spotify.PlaylistTrack {
	return s.tracks
}

// GetCurrentPlaylists returns the currently loaded playlists
func (s *SpotifyState) GetCurrentPlaylists() []spotify.SimplePlaylist {
	return s.playlists
}

// GetSelectedPlaylistID returns the ID of the currently selected playlist
func (s *SpotifyState) GetSelectedPlaylistID() string {
	return s.selectedPlaylistID
}
