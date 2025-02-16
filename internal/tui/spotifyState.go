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
}

type PlaylistsUpdatedMsg struct {
	Playlists []spotify.SimplePlaylist
	Err       error
}

type PlaylistSelectedMsg struct {
	PlaylistID string
}

type TracksUpdatedMsg struct {
	Tracks []spotify.PlaylistItem
	Err    error
}

type PlayerStateUpdatedMsg struct {
	State spotify.PlayerState
	Err   error
}

func NewSpotifyState(client *spotify.Client) *SpotifyState {
	log.Printf("Creating new SpotifyState with client: %v", client != nil)
	return &SpotifyState{
		client:             client,
		playlistItemsCache: make(map[string]*spotify.PlaylistItemPage),
	}
}

// TODO: This contains pagination at some point we likely need to refetch multiple pages
// FetchPlaylists retrieves all playlists for the current user and emits a state update
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

				Err:       err,
				Playlists: []spotify.SimplePlaylist{},
			}
		}

		s.playlists = playlists.Playlists
		log.Printf("SpotifyState: Successfully fetched %d playlists", len(s.playlists))
		return PlaylistsUpdatedMsg{
			Err:       nil,
			Playlists: s.playlists,
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
				Err:    fmt.Errorf("invalid playlist ID"),
				Tracks: []spotify.PlaylistItem{},
			}
		}

		// Check cache first
		if cachedItems, exists := s.playlistItemsCache[playlistID]; exists {
			log.Printf("SpotifyState: Found cached items for playlist %s", playlistID)
			s.tracks = cachedItems.Items
			return TracksUpdatedMsg{
				Err:    nil,
				Tracks: cachedItems.Items,
			}
		}

		// If not in cache, fetch from API
		log.Printf("SpotifyState: No cache found, fetching from API for playlist %s", playlistID)
		playlistItems, err := s.client.GetPlaylistItems(context.TODO(), spotify.ID(playlistID))
		if err != nil {
			log.Printf("SpotifyState: Error fetching playlist items: %v", err)
			return TracksUpdatedMsg{
				Err:    err,
				Tracks: []spotify.PlaylistItem{},
			}
		}

		// Cache the results
		s.playlistItemsCache[playlistID] = playlistItems
		s.tracks = playlistItems.Items
		log.Printf("SpotifyState: Successfully fetched and cached %d tracks for playlist %s", len(playlistItems.Items), playlistID)

		return TracksUpdatedMsg{
			Err:    nil,
			Tracks: playlistItems.Items,
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

// Get playback state
func (s *SpotifyState) FetchPlaybackState() tea.Cmd {
	return func() tea.Msg {
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err:   fmt.Errorf("failed to get playback state"),
				State: spotify.PlayerState{},
			}
		}
		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state
		return PlayerStateUpdatedMsg{
			Err:   nil,
			State: *state,
		}
	}
}

func (s *SpotifyState) StartPlayback() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Play(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error starting playback: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)

		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				State: spotify.PlayerState{},
				Err:   fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state

		return PlayerStateUpdatedMsg{
			State: *state,
			Err:   nil,
		}
	}
}

func (s *SpotifyState) PausePlayback() tea.Cmd {

	return func() tea.Msg {

		if err := s.client.Pause(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error pausing playback: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)

		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				State: spotify.PlayerState{},
				Err:   fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state

		return PlayerStateUpdatedMsg{
			State: *state,
			Err:   nil,
		}
	}
}

func (s *SpotifyState) NextTrack() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Next(context.TODO()); err != nil {
			log.Printf("SpotifyState: Error skipping to next track: %v", err)
			return nil
		}
		//TODO: Implement exponential backoff
		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				State: spotify.PlayerState{},
				Err:   fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state

		return PlayerStateUpdatedMsg{
			State: *state,
			Err:   nil,
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
		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				State: spotify.PlayerState{},
				Err:   fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state

		return PlayerStateUpdatedMsg{
			State: *state,
			Err:   nil,
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

		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				State: spotify.PlayerState{},
				Err:   fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state
		return PlayerStateUpdatedMsg{
			State: *state,
			Err:   nil,
		}
	}
}

func (s *SpotifyState) ToggleShuffleMode() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Shuffle(context.TODO(), !s.playerState.ShuffleState); err != nil {
			log.Printf("SpotifyState: Error toggling shuffle mode: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				State: spotify.PlayerState{},
				Err:   fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.playerState = *state
		return PlayerStateUpdatedMsg{
			State: *state,
			Err:   nil,
		}
	}
}

func (s *SpotifyState) ToggleRepeatMode() tea.Cmd {
	return func() tea.Msg {
		// Determine new repeat state based on current state
		var newState string
		if s.playerState.RepeatState == "off" {
			newState = "context"
		} else {
			newState = "off"
		}

		if err := s.client.Repeat(context.TODO(), newState); err != nil {
			log.Printf("SpotifyState: Error setting repeat mode: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				State: spotify.PlayerState{},
				Err:   fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)
		s.playerState = *state
		return PlayerStateUpdatedMsg{
			State: *state,
			Err:   nil,
		}
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
