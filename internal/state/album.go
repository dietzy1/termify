package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

func (s *SpotifyState) FetchAlbumTracks(ctx context.Context, albumId spotify.ID) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Fetching tracks for album: %s", albumId)
		if albumId == "" {
			log.Printf("SpotifyState: Invalid album ID")
			return ErrorMsg{
				Title:   "Invalid Input",
				Message: "Invalid album ID provided.",
			}
		}

		s.mu.RLock()
		cachedEntry, exists := s.tracksCache[albumId]
		s.mu.RUnlock()

		if exists {
			log.Printf("SpotifyState: Found cached tracks for album %s (%d tracks)",
				albumId, len(cachedEntry.Tracks))

			s.mu.Lock()
			s.tracks = make([]spotify.SimpleTrack, len(cachedEntry.Tracks))
			copy(s.tracks, cachedEntry.Tracks)
			s.mu.Unlock()

			return TracksUpdatedMsg{
				SourceID: albumId,
				Tracks:   cachedEntry.Tracks,
				NextPage: nil,
			}
		}

		log.Printf("SpotifyState: No cache found, fetching from API for album %s", albumId)
		albumTracks, err := s.client.GetAlbum(ctx, albumId)
		if err != nil {
			log.Printf("SpotifyState: Error fetching album info: %v", err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Fetch Album Info for %s", albumId),
				Message: err.Error(),
			}
		}

		allTracks := albumTracks.Tracks.Tracks

		for page := 1; ; page++ {
			log.Printf("SpotifyState: Fetching page %d of album tracks", page)
			err = s.client.NextPage(ctx, &albumTracks.Tracks)
			if err == spotify.ErrNoMorePages {
				break
			}
			if err != nil {
				log.Printf("SpotifyState: Error fetching next page of album tracks: %v", err)
				break
			}

			allTracks = append(allTracks, albumTracks.Tracks.Tracks...)
		}

		simpleTracks := make([]spotify.SimpleTrack, 0, len(allTracks))
		for _, item := range allTracks {
			item.Album.Name = albumTracks.Name
			simpleTracks = append(simpleTracks, item)
		}

		s.updateCacheEntry(albumId, simpleTracks, nil, false, len(simpleTracks), false)

		log.Printf("SpotifyState: Successfully fetched and cached %d tracks for album %s", len(simpleTracks), albumId)

		return TracksUpdatedMsg{
			SourceID: albumId,
			Tracks:   simpleTracks,
			NextPage: nil,
		}
	}
}
