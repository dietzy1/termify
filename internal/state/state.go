package state

import (
	"log"
	"sync"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type ErrorMsg struct {
	Title   string
	Message string
}

type PlaylistsUpdatedMsg struct {
}

type PlaylistSelectedMsg struct {
	PlaylistID string
}

type TracksUpdatedMsg struct {
}

type PlayerStateUpdatedMsg struct {
}

// SpotifyState manages all Spotify-related state and API calls
type SpotifyState struct {
	client *spotify.Client
	mu     sync.RWMutex

	deviceState []spotify.PlayerDevice
	playerState spotify.PlayerState

	// Cache for current data
	playlists []spotify.SimplePlaylist
	tracks    []spotify.SimpleTrack

	// Cache map for tracks by source ID (playlist, album, artist)
	tracksCache map[spotify.ID][]spotify.SimpleTrack

	searchResults struct {
		tracks    []spotify.FullTrack
		artists   []spotify.FullArtist
		albums    []spotify.SimpleAlbum
		playlists []spotify.SimplePlaylist
	}

	Queue Queue

	selectedID spotify.ID
}

func NewSpotifyState(client *spotify.Client) *SpotifyState {
	log.Printf("Creating new SpotifyState with client: %v", client != nil)
	return &SpotifyState{
		client:      client,
		mu:          sync.RWMutex{},
		tracksCache: make(map[spotify.ID][]spotify.SimpleTrack),
	}
}

// Function which logs the content of the OATH token in the client
func (s *SpotifyState) GetOathToken() *oauth2.Token {
	token, err := s.client.Token()
	if err != nil {
		log.Printf("Error getting token: %v", err)
		return nil
	}
	return token
}

func (s *SpotifyState) GetDeviceState() []spotify.PlayerDevice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy to prevent concurrent modification
	if s.deviceState == nil {
		return nil
	}
	deviceCopy := make([]spotify.PlayerDevice, len(s.deviceState))
	copy(deviceCopy, s.deviceState)
	return deviceCopy
}

func (s *SpotifyState) GetPlayerState() spotify.PlayerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.playerState
}

func (s *SpotifyState) GetPlaylists() []spotify.SimplePlaylist {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.playlists == nil {
		return nil
	}
	playlistsCopy := make([]spotify.SimplePlaylist, len(s.playlists))
	copy(playlistsCopy, s.playlists)
	return playlistsCopy
}

func (s *SpotifyState) GetTracks() []spotify.SimpleTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.tracks == nil {
		return nil
	}
	tracksCopy := make([]spotify.SimpleTrack, len(s.tracks))
	copy(tracksCopy, s.tracks)
	return tracksCopy
}

func (s *SpotifyState) GetTracksCached(sourceID spotify.ID) ([]spotify.SimpleTrack, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tracks, exists := s.tracksCache[sourceID]
	if !exists {
		return nil, false
	}

	tracksCopy := make([]spotify.SimpleTrack, len(tracks))
	copy(tracksCopy, tracks)
	return tracksCopy, true
}

func (s *SpotifyState) GetSearchResultTracks() []spotify.FullTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.searchResults.tracks == nil {
		return nil
	}
	tracksCopy := make([]spotify.FullTrack, len(s.searchResults.tracks))
	copy(tracksCopy, s.searchResults.tracks)
	return tracksCopy
}

func (s *SpotifyState) GetSearchResultArtists() []spotify.FullArtist {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.searchResults.artists == nil {
		return nil
	}
	artistsCopy := make([]spotify.FullArtist, len(s.searchResults.artists))
	copy(artistsCopy, s.searchResults.artists)
	return artistsCopy
}

func (s *SpotifyState) GetSearchResultAlbums() []spotify.SimpleAlbum {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.searchResults.albums == nil {
		return nil
	}
	albumsCopy := make([]spotify.SimpleAlbum, len(s.searchResults.albums))
	copy(albumsCopy, s.searchResults.albums)
	return albumsCopy
}

func (s *SpotifyState) GetSearchResultPlaylists() []spotify.SimplePlaylist {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.searchResults.playlists == nil {
		return nil
	}
	playlistsCopy := make([]spotify.SimplePlaylist, len(s.searchResults.playlists))
	copy(playlistsCopy, s.searchResults.playlists)
	return playlistsCopy
}

func (s *SpotifyState) GetSelectedID() spotify.ID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedID
}

func (s *SpotifyState) SetDeviceState(devices []spotify.PlayerDevice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deviceState = devices
}

func (s *SpotifyState) SetPlayerState(state spotify.PlayerState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerState = state
}

func (s *SpotifyState) SetPlaylists(playlists []spotify.SimplePlaylist) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playlists = playlists
}

func (s *SpotifyState) SetTracks(tracks []spotify.SimpleTrack) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracks = tracks
}

func (s *SpotifyState) AddToTracksCache(sourceID spotify.ID, tracks []spotify.SimpleTrack) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracksCache[sourceID] = tracks
}

func (s *SpotifyState) ClearTracksCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracksCache = make(map[spotify.ID][]spotify.SimpleTrack)
}

func (s *SpotifyState) setSearchResults(tracks []spotify.FullTrack, artists []spotify.FullArtist,
	albums []spotify.SimpleAlbum, playlists []spotify.SimplePlaylist) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchResults.tracks = tracks
	s.searchResults.artists = artists
	s.searchResults.albums = albums
	s.searchResults.playlists = playlists
}

func (s *SpotifyState) SetSelectedID(id spotify.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedID = id
}
