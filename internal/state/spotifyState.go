package state

import (
	"context"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

//TODO: pass a context into all the API functions like a human its fine if its a shared context from the TUI I think

// SpotifyState manages all Spotify-related state and API calls
type SpotifyState struct {
	client *spotify.Client

	DeviceState spotify.PlayerDevice
	//Currently playing state
	PlayerState spotify.PlayerState

	// Cache for current data
	Playlists []spotify.SimplePlaylist
	Tracks    []spotify.PlaylistItem

	// Cache map for playlist items
	playlistItemsCache map[string]*spotify.PlaylistItemPage

	// Currently selected items
	selectedPlaylistID string
}

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

func NewSpotifyState(client *spotify.Client) *SpotifyState {
	log.Printf("Creating new SpotifyState with client: %v", client != nil)
	return &SpotifyState{
		client:             client,
		playlistItemsCache: make(map[string]*spotify.PlaylistItemPage),
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
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.PlayerState = *state

		return PlayerStateUpdatedMsg{
			Err: nil,
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
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.PlayerState = *state

		return PlayerStateUpdatedMsg{
			Err: nil,
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
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.PlayerState = *state
		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) ToggleShuffleMode() tea.Cmd {
	return func() tea.Msg {
		if err := s.client.Shuffle(context.TODO(), !s.PlayerState.ShuffleState); err != nil {
			log.Printf("SpotifyState: Error toggling shuffle mode: %v", err)
			return nil
		}

		time.Sleep(500 * time.Millisecond)
		// We can do this alot smarter by checking if the track has changed
		state, err := s.client.PlayerState(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching playback state: %v", err)
			return PlayerStateUpdatedMsg{
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)

		s.PlayerState = *state
		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}

func (s *SpotifyState) ToggleRepeatMode() tea.Cmd {
	return func() tea.Msg {
		// Determine new repeat state based on current state
		var newState string
		if s.PlayerState.RepeatState == "off" {
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
				Err: fmt.Errorf("invalid playlist ID"),
			}
		}

		log.Println("SpotifyState: Player state:", state)
		s.PlayerState = *state
		return PlayerStateUpdatedMsg{
			Err: nil,
		}
	}
}

// We need to look at current playback and check if isPlaying is true firstly
// We then need to compare progressMS  with durationMs and if it is equal then we need to refetch the current playback state and update the UI

//TODO: this is actually important this if no device is active then we cannot play anything
// TODO: We need to figure out if we also want to go with the librespot daemon way and perhabs use interfaces to define similar behaviour.

func (s *SpotifyState) FetchDevices() tea.Cmd {
	return func() tea.Msg {
		devices, err := s.client.PlayerDevices(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching player devices: %v", err)
			return nil
		}

		if len(devices) == 0 {
			log.Println("SpotifyState: No devices found")
			return nil
		}

		for _, device := range devices {
			log.Printf("SpotifyState: Found device: %v", device.Name)
		}
		// This is unsafe and bad TODO: fix this later
		s.DeviceState = devices[0]
		return nil
	}
}

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
