package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

func (s *SpotifyState) FetchTopTracks(ctx context.Context, artistId spotify.ID) tea.Cmd {
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
		cachedEntry, exists := s.tracksCache[artistId]
		s.mu.RUnlock()

		if exists {
			log.Printf("SpotifyState: Found cached tracks for artist %s (%d tracks)",
				artistId, len(cachedEntry.Tracks))

			s.mu.Lock()
			s.tracks = make([]spotify.SimpleTrack, len(cachedEntry.Tracks))
			copy(s.tracks, cachedEntry.Tracks)
			s.mu.Unlock()

			return TracksUpdatedMsg{
				SourceID: artistId,
				Tracks:   cachedEntry.Tracks,
				NextPage: nil,
			}
		}

		topTracks, err := s.client.GetArtistsTopTracks(ctx, artistId, "US")
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

		s.updateCacheEntry(artistId, simpleTracks, nil, false, len(simpleTracks), 0, false)

		log.Printf("SpotifyState: Successfully fetched and cached %d top tracks for artist %s", len(simpleTracks), artistId)

		return TracksUpdatedMsg{
			SourceID: artistId,
			Tracks:   simpleTracks,
			NextPage: nil,
		}
	}
}
