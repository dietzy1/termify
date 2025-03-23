package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

func (s *SpotifyState) FetchAlbumTracks(albumId spotify.ID) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Fetching tracks for album: %s", albumId)
		if albumId == "" {
			log.Printf("SpotifyState: Invalid album ID")
			return TracksUpdatedMsg{
				Err: fmt.Errorf("invalid album ID"),
			}
		}

		if cachedTracks, exists := s.tracksCache[albumId]; exists {
			log.Printf("SpotifyState: Found cached tracks for album %s", albumId)
			s.Tracks = cachedTracks
			log.Printf("SpotifyState: Successfully loaded %d tracks from cache for album %s", len(cachedTracks), albumId)
			return TracksUpdatedMsg{
				Err: nil,
			}
		}

		albumTracks, err := s.client.GetAlbum(context.TODO(), albumId)
		if err != nil {
			log.Printf("SpotifyState: Error fetching album tracks: %v", err)
			return TracksUpdatedMsg{
				Err: err,
			}
		}

		allTracks := albumTracks.Tracks.Tracks

		for page := 1; ; page++ {
			log.Printf("SpotifyState: Fetching page %d of playlist items", page)
			err = s.client.NextPage(context.TODO(), &albumTracks.Tracks)
			if err == spotify.ErrNoMorePages {
				break
			}
			if err != nil {
				log.Println("SpotifyState: Error fetching next page of playlist items:", err)
				break
			}

			allTracks = append(allTracks, albumTracks.Tracks.Tracks...)
		}

		simpleTracks := make([]spotify.SimpleTrack, 0, len(allTracks))
		for _, item := range allTracks {
			item.Album.Name = albumTracks.Name
			simpleTracks = append(simpleTracks, item)
		}

		s.Tracks = simpleTracks
		s.tracksCache[albumId] = simpleTracks

		log.Printf("SpotifyState: Successfully fetched and cached %d tracks for album %s", len(allTracks), albumId)
		return TracksUpdatedMsg{
			Err: nil,
		}
	}
}
