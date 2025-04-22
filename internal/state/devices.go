package state

import (
	"context"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type DevicesUpdatedMsg struct {
	err error
}

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

		s.mu.Lock()
		s.deviceState = devices
		s.mu.Unlock()

		//TODO: Here we have an option of potentialling setting a setting as the default transfer playback.
		err = s.client.TransferPlayback(
			context.TODO(),
			devices[0].ID,
			false,
		)
		if err != nil {
			log.Println("Failed to transfer playback to index 0")
			return nil
		}

		return DevicesUpdatedMsg{
			err: nil,
		}
	}
}

func (s *SpotifyState) SelectDevice(deviceID spotify.ID) tea.Cmd {
	return func() tea.Msg {

		err := s.client.TransferPlayback(context.TODO(), deviceID, false)
		if err != nil {
			log.Printf("SpotifyState: Error transferring playback to device %v: %v", deviceID, err)
			return DevicesUpdatedMsg{
				err: err,
			}
		}

		log.Printf("Selected device %s", deviceID)
		time.Sleep(500 * time.Millisecond)

		devices, err := s.client.PlayerDevices(context.TODO())
		if err != nil {
			log.Printf("SpotifyState: Error fetching player devices: %v", err)
			return nil
		}

		s.mu.Lock()
		s.deviceState = devices
		s.mu.Unlock()

		return DevicesUpdatedMsg{
			err: nil,
		}
	}
}
