package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Status struct {
	Username       string  `json:"username"`
	DeviceID       string  `json:"device_id"`
	DeviceName     string  `json:"device_name"`
	PlayOrigin     string  `json:"play_origin"`
	Stopped        bool    `json:"stopped"`
	Paused         bool    `json:"paused"`
	Buffering      bool    `json:"buffering"`
	Volume         float64 `json:"volume"`
	VolumeSteps    float64 `json:"volume_steps"`
	RepeatContext  bool    `json:"repeat_context"`
	RepeatTrack    bool    `json:"repeat_track"`
	ShuffleContext bool    `json:"shuffle_context"`
	Track          Track   `json:"track"`
}

func (c *Client) GetStatus() (*Status, error) {
	resp, err := c.doRequest("GET", "/status", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var status Status
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}
