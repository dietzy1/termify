package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Metadata struct {
	URI           string   `json:"uri"`
	Name          string   `json:"name"`
	ArtistNames   []string `json:"artist_names"`
	AlbumName     string   `json:"album_name"`
	AlbumCoverURL string   `json:"album_cover_url"`
	Position      int      `json:"position"`
	Duration      int      `json:"duration"`
}

type PlaybackEvent struct {
	URI        string `json:"uri"`
	PlayOrigin string `json:"play_origin"`
}

type SeekEvent struct {
	URI        string `json:"uri"`
	Position   int    `json:"position"`
	Duration   int    `json:"duration"`
	PlayOrigin string `json:"play_origin"`
}

type VolumeEvent struct {
	Value int `json:"value"`
	Max   int `json:"max"`
}

type BooleanEvent struct {
	Value bool `json:"value"`
}

type WsClient struct {
	conn   *websocket.Conn
	Events chan Event
}

func NewClient(host string) (*WsClient, error) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/events"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	client := &WsClient{
		conn:   c,
		Events: make(chan Event),
	}

	go client.readEvents()

	return client, nil
}

func (c *WsClient) readEvents() {
	defer close(c.Events)
	for {
		var event Event
		err := c.conn.ReadJSON(&event)
		if err != nil {
			log.Println("read:", err)
			return
		}
		c.Events <- event
	}
}

func (c *WsClient) Close() error {
	return c.conn.Close()
}

func HandleEvent(event Event) {
	switch event.Type {
	case "active":
		fmt.Println("Device became active")
	case "inactive":
		fmt.Println("Device became inactive")
	case "metadata":
		var metadata Metadata
		json.Unmarshal(event.Payload, &metadata)
		fmt.Printf("New track loaded: %s by %s\n", metadata.Name, metadata.ArtistNames)
	case "will_play", "playing", "not_playing", "paused", "stopped":
		var playbackEvent PlaybackEvent
		json.Unmarshal(event.Payload, &playbackEvent)
		fmt.Printf("%s event: %s (Origin: %s)\n", event.Type, playbackEvent.URI, playbackEvent.PlayOrigin)
	case "seek":
		var seekEvent SeekEvent
		json.Unmarshal(event.Payload, &seekEvent)
		fmt.Printf("Seek event: %s at %d/%d ms (Origin: %s)\n", seekEvent.URI, seekEvent.Position, seekEvent.Duration, seekEvent.PlayOrigin)
	case "volume":
		var volumeEvent VolumeEvent
		json.Unmarshal(event.Payload, &volumeEvent)
		fmt.Printf("Volume changed: %d/%d\n", volumeEvent.Value, volumeEvent.Max)
	case "shuffle_context", "repeat_context", "repeat_track":
		var booleanEvent BooleanEvent
		json.Unmarshal(event.Payload, &booleanEvent)
		fmt.Printf("%s changed: %v\n", event.Type, booleanEvent.Value)
	default:
		fmt.Printf("Unknown event type: %s\n", event.Type)
	}
}
