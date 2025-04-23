package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

func (s *SpotifyState) FetchTopTracks(artistId spotify.ID) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Fetching top tracks for artist: %s", artistId)
		if artistId == "" {
			log.Printf("SpotifyState: Invalid artist ID")
			return ErrorMsg{
				Title:   "Invalid Input",
				Message: "Invalid artist ID provided.",
			}
		}

		s.mu.RLock()
		cachedTracks, exists := s.tracksCache[artistId]
		s.mu.RUnlock()

		if exists {
			log.Printf("SpotifyState: Found cached top tracks for artist %s", artistId)
			s.mu.Lock()
			s.tracks = cachedTracks
			s.mu.Unlock()

			log.Printf("SpotifyState: Successfully loaded %d tracks from cache for artist %s", len(cachedTracks), artistId)
			return TracksUpdatedMsg{}
		}

		topTracks, err := s.client.GetArtistsTopTracks(context.TODO(), artistId, "US")
		if err != nil {
			log.Printf("SpotifyState: Error fetching top tracks for artist %s: %v", artistId, err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Fetch Top Tracks for Artist %s", artistId),
				Message: err.Error(),
			}
		}

		simpleTracks := make([]spotify.SimpleTrack, 0, len(topTracks))
		for _, track := range topTracks {
			track.SimpleTrack.Album = track.Album
			simpleTracks = append(simpleTracks, track.SimpleTrack)
		}

		s.mu.Lock()
		s.tracks = simpleTracks
		s.tracksCache[artistId] = simpleTracks
		s.mu.Unlock()

		log.Printf("SpotifyState: Successfully fetched and cached %d top tracks for artist %s", len(simpleTracks), artistId)
		return TracksUpdatedMsg{}
	}
}
