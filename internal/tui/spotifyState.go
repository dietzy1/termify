package tui

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

// SpotifyState manages all Spotify-related state and API calls
type SpotifyState struct {
	client *spotify.Client

	//Currently playing state
	playerState spotify.PlayerState

	// Cache for current data
	playlists []spotify.SimplePlaylist
	tracks    []spotify.PlaylistItem

	// Cache map for playlist items
	playlistItemsCache map[string]*spotify.PlaylistItemPage

	// Currently selected items
	selectedPlaylistID string
	selectedTrackID    string
}

// Message types for state updates
type StateUpdateMsg struct {
	Type StateUpdateType
	Data interface{}
	Err  error
}

type StateUpdateType int

const (
	PlaylistsUpdated StateUpdateType = iota
	TracksUpdated
	PlaylistSelected
	TrackSelected
	PlayerStateUpdated
)

func NewSpotifyState(client *spotify.Client) *SpotifyState {
	log.Printf("Creating new SpotifyState with client: %v", client != nil)
	return &SpotifyState{
		client:             client,
		playlistItemsCache: make(map[string]*spotify.PlaylistItemPage),
	}
}

// FetchPlaylists retrieves all playlists for the current user and emits a state update
func (s *SpotifyState) FetchPlaylists() tea.Cmd {
	log.Printf("SpotifyState: Starting FetchPlaylists command creation")
	return func() tea.Msg {
		log.Printf("SpotifyState: Executing FetchPlaylists, client: %v", s.client != nil)
		if s.client == nil {
			log.Printf("SpotifyState: Error - client is nil")
			return StateUpdateMsg{
				Type: PlaylistsUpdated,
				Err:  fmt.Errorf("spotify client not initialized"),
			}
		}

		playlists, err := s.client.CurrentUsersPlaylists(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playlists: %v", err)
			return StateUpdateMsg{
				Type: PlaylistsUpdated,
				Err:  err,
			}
		}

		s.playlists = playlists.Playlists
		log.Printf("SpotifyState: Successfully fetched %d playlists", len(s.playlists))
		return StateUpdateMsg{
			Type: PlaylistsUpdated,
			Data: s.playlists,
		}
	}
}

// FetchPlaylistTracks retrieves all tracks for a given playlist and emits a state update
func (s *SpotifyState) FetchPlaylistTracks(playlistID string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Fetching tracks for playlist: %s", playlistID)
		if playlistID == "" {
			log.Printf("SpotifyState: Invalid playlist ID")
			return StateUpdateMsg{
				Type: TracksUpdated,
				Err:  fmt.Errorf("invalid playlist ID"),
			}
		}

		// Check cache first
		if cachedItems, exists := s.playlistItemsCache[playlistID]; exists {
			log.Printf("SpotifyState: Found cached items for playlist %s", playlistID)
			s.tracks = cachedItems.Items
			return StateUpdateMsg{
				Type: TracksUpdated,
				Data: s.tracks,
			}
		}

		// If not in cache, fetch from API
		log.Printf("SpotifyState: No cache found, fetching from API for playlist %s", playlistID)
		playlistItems, err := s.client.GetPlaylistItems(context.TODO(), spotify.ID(playlistID))
		if err != nil {
			log.Printf("SpotifyState: Error fetching playlist items: %v", err)
			return StateUpdateMsg{
				Type: TracksUpdated,
				Err:  err,
			}
		}

		// Cache the results
		s.playlistItemsCache[playlistID] = playlistItems
		s.tracks = playlistItems.Items
		log.Printf("SpotifyState: Successfully fetched and cached %d tracks for playlist %s", len(playlistItems.Items), playlistID)

		return StateUpdateMsg{
			Type: TracksUpdated,
			Data: s.tracks,
		}
	}
}

func (s *SpotifyState) SelectPlaylist(playlistID string) tea.Cmd {
	return func() tea.Msg {

		playlistID = strings.Split(playlistID, ":")[2]
		s.selectedPlaylistID = playlistID
		return StateUpdateMsg{
			Type: PlaylistSelected,
			Data: playlistID,
		}
	}
}

// Get playback state
func (s *SpotifyState) GetPlaybackState() tea.Cmd {
	return func() tea.Msg {
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return StateUpdateMsg{
				Type: PlayerStateUpdated,
				Err:  fmt.Errorf("invalid playlist ID"),
			}
		}
		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state

		return StateUpdateMsg{
			Type: PlayerStateUpdated,
			Data: state,
			Err:  nil,
		}
	}
}

func (s *SpotifyState) StartPlayback() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Play(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error starting playback: %v", err)
			return nil
		}

		return StateUpdateMsg{}
	}
}

func (s *SpotifyState) PausePlayback() tea.Cmd {

	return func() tea.Msg {

		if err := s.client.Pause(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error pausing playback: %v", err)
			return nil
		}

		return StateUpdateMsg{}
	}
}

func (s *SpotifyState) NextTrack() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Next(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error skipping to next track: %v", err)
			return nil
		}
		//TODO: Implement exponential backoff
		time.Sleep(200 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return StateUpdateMsg{
				Type: PlayerStateUpdated,
				Err:  fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state

		return StateUpdateMsg{
			Type: PlayerStateUpdated,
			Data: state,
			Err:  nil,
		}

	}
}

func (s *SpotifyState) PreviousTrack() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Previous(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error skipping to previous track: %v", err)
			return nil
		}

		//TODO: Implement exponential backoff
		time.Sleep(200 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return StateUpdateMsg{
				Type: PlayerStateUpdated,
				Err:  fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state

		return StateUpdateMsg{
			Type: PlayerStateUpdated,
			Data: state,
			Err:  nil,
		}
	}
}

func (s *SpotifyState) PlayTrack(trackID spotify.ID) tea.Cmd {
	return func() tea.Msg {
		if err := s.client.PlayOpt(context.TODO(), &spotify.PlayOptions{
			URIs: []spotify.URI{"spotify:track:" + spotify.URI(trackID)},
		}); err != nil {
			log.Printf("SpotifyState: Error playing track: %v", err)
			return nil
		}
		return StateUpdateMsg{}
	}
}

// We need to look at current playback and check if isPlaying is true firstly
// We then need to compare progressMS  with durationMs and if it is equal then we need to refetch the current playback state and update the UI

//TODO: this is actually important this if no device is active then we cannot play anything
// TODO: We need to figure out if we also want to go with the librespot daemon way and perhabs use interfaces to define similar behaviour.
/* func (s *SpotifyState) TransferPlaybackToTermify() tea.Cmd {

	return func() tea.Msg {
		devices, err := s.client.PlayerDevices(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching player devices: %v", err)
			return nil
		}
		log.Println("SpotifyState: Found devices:", devices)

		s.client.TransferPlayback(
			context.TODO(),
			devices[0].ID,
			false,
		)
		return StateUpdateMsg{}
	}
} */

// GetCurrentTracks returns the currently loaded tracks
/* func (s *SpotifyState) GetCurrentTracks() []spotify.PlaylistItem {
	return s.tracks
}

// GetCurrentPlaylists returns the currently loaded playlists
func (s *SpotifyState) GetCurrentPlaylists() []spotify.SimplePlaylist {
	return s.playlists
}

// GetSelectedPlaylistID returns the ID of the currently selected playlist
func (s *SpotifyState) GetSelectedPlaylistID() string {
	return s.selectedPlaylistID
} */

// SelectPlaylist updates the selected playlist and emits a state update
