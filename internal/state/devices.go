package state

import (
	"context"
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type DevicesUpdatedMsg struct{}

func (s *SpotifyState) FetchDevices(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		devices, err := s.client.PlayerDevices(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching player devices: %v", err)
			return ErrorMsg{
				Title:   "Failed to Fetch Devices",
				Message: err.Error(),
			}
		}

		if len(devices) == 0 {
			log.Println("SpotifyState: No devices found")
			return ErrorMsg{
				Title:   "No Spotify Devices Found",
				Message: "Could not find any active Spotify devices. Please ensure Spotify is running on a device.",
			}
		}

		for _, device := range devices {
			log.Printf("SpotifyState: Found device: %v", device.Name)
		}

		activeDevice := -1
		for i, device := range devices {
			if device.Active {
				activeDevice = i
				break
			}
		}

		if activeDevice == -1 && len(devices) > 0 {
			log.Printf("SpotifyState: No active device found, attempting to transfer playback to %s", devices[0].Name)
			err = s.client.TransferPlayback(ctx, devices[0].ID, false)
			if err != nil {
				log.Printf("SpotifyState: Failed to transfer playback to device %s: %v", devices[0].Name, err)
				return ErrorMsg{
					Title:   fmt.Sprintf("Failed to Activate Device %s", devices[0].Name),
					Message: err.Error(),
				}
			}
			time.Sleep(500 * time.Millisecond)
			devices, err = s.client.PlayerDevices(ctx)
			if err != nil {
				log.Printf("SpotifyState: Error re-fetching devices after transfer attempt: %v", err)
				return ErrorMsg{
					Title:   "Failed to Refresh Devices",
					Message: fmt.Sprintf("Attempted to activate device %s, but failed to refresh device list: %s", devices[0].Name, err.Error()),
				}
			}
		} else if activeDevice != -1 {
			log.Printf("SpotifyState: Active device found: %s", devices[activeDevice].Name)
		}

		s.mu.Lock()
		s.deviceState = devices
		s.mu.Unlock()

		return DevicesUpdatedMsg{}
	}
}

func (s *SpotifyState) SelectDevice(ctx context.Context, deviceID spotify.ID) tea.Cmd {
	return func() tea.Msg {

		err := s.client.TransferPlayback(ctx, deviceID, false)
		if err != nil {
			log.Printf("SpotifyState: Error transferring playback to device %v: %v", deviceID, err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Select Device %s", deviceID),
				Message: err.Error(),
			}
		}

		log.Printf("Selected device %s", deviceID)
		time.Sleep(500 * time.Millisecond)

		devices, err := s.client.PlayerDevices(ctx)
		if err != nil {
			log.Printf("SpotifyState: Error fetching player devices after selecting %s: %v", deviceID, err)
			return ErrorMsg{
				Title:   "Failed to Refresh Devices",
				Message: fmt.Sprintf("Selected device %s, but failed to refresh device list: %s", deviceID, err.Error()),
			}
		}

		s.mu.Lock()
		s.deviceState = devices
		s.mu.Unlock()

		return DevicesUpdatedMsg{}
	}
}
