package state

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

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
		s.mu.Lock()
		s.deviceState = devices
		s.mu.Unlock()

		err = s.client.TransferPlayback(
			context.TODO(),
			devices[0].ID,
			false,
		)

		return nil
	}
}
