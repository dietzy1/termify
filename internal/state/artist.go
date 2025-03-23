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
			return TracksUpdatedMsg{
				Err: fmt.Errorf("invalid artist ID"),
			}
		}

		if cachedTracks, exists := s.tracksCache[artistId]; exists {
			log.Printf("SpotifyState: Found cached top tracks for artist %s", artistId)
			s.Tracks = cachedTracks
			log.Printf("SpotifyState: Successfully loaded %d tracks from cache for artist %s", len(cachedTracks), artistId)
			return TracksUpdatedMsg{
				Err: nil,
			}
		}

		topTracks, err := s.client.GetArtistsTopTracks(context.TODO(), artistId, "US")
		if err != nil {
			log.Printf("SpotifyState: Error fetching top tracks: %v", err)
			return TracksUpdatedMsg{
				Err: err,
			}
		}

		simpleTracks := make([]spotify.SimpleTrack, 0, len(topTracks))
		for _, track := range topTracks {
			track.SimpleTrack.Album = track.Album
			simpleTracks = append(simpleTracks, track.SimpleTrack)
		}

		s.Tracks = simpleTracks
		s.tracksCache[artistId] = simpleTracks

		log.Printf("SpotifyState: Successfully fetched and cached %d top tracks for artist %s", len(simpleTracks), artistId)
		return TracksUpdatedMsg{
			Err: nil,
		}
	}
}
