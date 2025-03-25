package state

import (
	"log"
	"sync"

	"github.com/zmb3/spotify/v2"
)

type PlaylistsUpdatedMsg struct {
	Err error
}

type PlaylistSelectedMsg struct {
	PlaylistID string
}

type TracksUpdatedMsg struct {
	Err error
}

type PlayerStateUpdatedMsg struct {
	Err error
}

//TODO: pass a context into all the API functions like a human its fine if its a shared context from the TUI I think

// SpotifyState manages all Spotify-related state and API calls
type SpotifyState struct {
	client *spotify.Client
	mu     sync.RWMutex

	deviceState []spotify.PlayerDevice
	//Currently playing state
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

	// Queue of tracks to play
	queue []spotify.FullTrack

	// Currently selected items
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

// IsQueueEmpty returns true if the queue is empty or nil
func (s *SpotifyState) IsQueueEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.queue == nil || len(s.queue) == 0
}

// GetDeviceState returns a copy of the device state
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

// GetPlayerState returns a copy of the player state
func (s *SpotifyState) GetPlayerState() spotify.PlayerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.playerState
}

// GetPlaylists returns a copy of the playlists
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

// GetTracks returns a copy of the tracks
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

// GetTracksCached returns a copy of the tracks for a specific source ID from cache
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

// GetSearchResultTracks returns a copy of the search result tracks
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

// GetSearchResultArtists returns a copy of the search result artists
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

// GetSearchResultAlbums returns a copy of the search result albums
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

// GetSearchResultPlaylists returns a copy of the search result playlists
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

// GetQueue returns a copy of the queue
func (s *SpotifyState) GetQueue() []spotify.FullTrack {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.queue == nil {
		return nil
	}
	queueCopy := make([]spotify.FullTrack, len(s.queue))
	copy(queueCopy, s.queue)
	return queueCopy
}

// GetSelectedID returns the currently selected ID
func (s *SpotifyState) GetSelectedID() spotify.ID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selectedID
}

// SetDeviceState safely updates the device state
func (s *SpotifyState) SetDeviceState(devices []spotify.PlayerDevice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deviceState = devices
}

// SetPlayerState safely updates the player state
func (s *SpotifyState) SetPlayerState(state spotify.PlayerState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playerState = state
}

// SetPlaylists safely updates the playlists
func (s *SpotifyState) SetPlaylists(playlists []spotify.SimplePlaylist) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playlists = playlists
}

// SetTracks safely updates the tracks
func (s *SpotifyState) SetTracks(tracks []spotify.SimpleTrack) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracks = tracks
}

// AddToTracksCache safely adds tracks to the cache
func (s *SpotifyState) AddToTracksCache(sourceID spotify.ID, tracks []spotify.SimpleTrack) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracksCache[sourceID] = tracks
}

// ClearTracksCache safely clears the tracks cache
func (s *SpotifyState) ClearTracksCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tracksCache = make(map[spotify.ID][]spotify.SimpleTrack)
}

// SetSearchResults safely updates all search results
func (s *SpotifyState) SetSearchResults(tracks []spotify.FullTrack, artists []spotify.FullArtist,
	albums []spotify.SimpleAlbum, playlists []spotify.SimplePlaylist) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.searchResults.tracks = tracks
	s.searchResults.artists = artists
	s.searchResults.albums = albums
	s.searchResults.playlists = playlists
}

// SetQueue safely updates the queue
func (s *SpotifyState) SetQueue(queue []spotify.FullTrack) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = queue
}

// AddTrackToQueueLocal safely adds an item to the queue without API calls
func (s *SpotifyState) AddTrackToQueueLocal(track spotify.FullTrack) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = append(s.queue, track)
}

// ClearQueue safely clears the queue
func (s *SpotifyState) ClearQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queue = nil
}

// SetSelectedID safely updates the selected ID
func (s *SpotifyState) SetSelectedID(id spotify.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedID = id
}
