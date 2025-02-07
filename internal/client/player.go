package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Track struct {
	URI           string   `json:"uri"`
	Name          string   `json:"name"`
	ArtistNames   []string `json:"artist_names"`
	AlbumName     string   `json:"album_name"`
	AlbumCoverURL string   `json:"album_cover_url"`
	Duration      int      `json:"duration"`
}

type Play struct {
	URI       string `json:"uri"`
	SkipToURI string `json:"skip_to_uri,omitempty"`
	Paused    bool   `json:"paused,omitempty"`
}

type VolumeInfo struct {
	Value int `json:"value"`
	Max   int `json:"max"`
}

func (c *Client) Play(play Play) error {
	resp, err := c.doRequest("POST", "/player/play", play)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Resume() error {
	resp, err := c.doRequest("POST", "/player/resume", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Pause() error {
	resp, err := c.doRequest("POST", "/player/pause", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) PlayPause() error {
	resp, err := c.doRequest("POST", "/player/playpause", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Next() error {
	resp, err := c.doRequest("POST", "/player/next", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Previous() error {
	resp, err := c.doRequest("POST", "/player/prev", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Seek(position int, relative bool) error {
	body := struct {
		Position int  `json:"position"`
		Relative bool `json:"relative"`
	}{
		Position: position,
		Relative: relative,
	}

	resp, err := c.doRequest("POST", "/player/seek", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetVolume() (*VolumeInfo, error) {
	resp, err := c.doRequest("GET", "/player/volume", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var volumeInfo VolumeInfo
	err = json.NewDecoder(resp.Body).Decode(&volumeInfo)
	if err != nil {
		return nil, err
	}

	return &volumeInfo, nil
}

func (c *Client) SetVolume(volume float64, relative bool) error {
	body := struct {
		Volume   float64 `json:"volume"`
		Relative bool    `json:"relative"`
	}{
		Volume:   volume,
		Relative: relative,
	}

	resp, err := c.doRequest("POST", "/player/volume", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) SetRepeatContext(enabled bool) error {
	body := struct {
		RepeatContext bool `json:"repeat_context"`
	}{
		RepeatContext: enabled,
	}

	resp, err := c.doRequest("POST", "/player/repeat_context", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) SetRepeatTrack(enabled bool) error {
	body := struct {
		RepeatTrack bool `json:"repeat_track"`
	}{
		RepeatTrack: enabled,
	}

	resp, err := c.doRequest("POST", "/player/repeat_track", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) SetShuffleContext(enabled bool) error {
	body := struct {
		ShuffleContext bool `json:"shuffle_context"`
	}{
		ShuffleContext: enabled,
	}

	resp, err := c.doRequest("POST", "/player/shuffle_context", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) AddToQueue(uri string) error {
	body := struct {
		URI string `json:"uri"`
	}{
		URI: uri,
	}

	resp, err := c.doRequest("POST", "/player/add_to_queue", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
