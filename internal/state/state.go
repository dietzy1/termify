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

	fetchingPages map[string]bool
	tracksCache   map[spotify.ID]*CacheEntry

	searchResults struct {
		tracks    []spotify.FullTrack
		artists   []spotify.FullArtist
		albums    []spotify.SimpleAlbum
		playlists []spotify.SimplePlaylist
	}

	Queue QueueManager

	selectedID spotify.ID
}

func NewSpotifyState(client *spotify.Client) *SpotifyState {
	log.Printf("Creating new SpotifyState with client: %v", client != nil)
	return &SpotifyState{
		client:        client,
		mu:            sync.RWMutex{},
		tracksCache:   make(map[spotify.ID]*CacheEntry),
		fetchingPages: make(map[string]bool),
	}
}

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
