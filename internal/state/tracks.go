package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type TracksUpdatedMsg struct {
	SourceID spotify.ID
	Tracks   []spotify.SimpleTrack
	NextPage *spotify.PlaylistItemPage
}

func (s *SpotifyState) FetchPlaylistTracks(ctx context.Context, playlistID spotify.ID) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Fetching tracks for playlist: %s", playlistID)
		if playlistID == "" {
			log.Printf("SpotifyState: Invalid playlist ID")
			return ErrorMsg{
				Title:   "Invalid Input",
				Message: "Invalid playlist ID provided",
			}
		}

		s.mu.RLock()
		cachedEntry, exists := s.tracksCache[playlistID]
		s.mu.RUnlock()

		if exists {
			log.Printf("SpotifyState: Found cached tracks for playlist %s (%d tracks)",
				playlistID, len(cachedEntry.Tracks))

			s.mu.Lock()
			s.tracks = make([]spotify.SimpleTrack, len(cachedEntry.Tracks))
			copy(s.tracks, cachedEntry.Tracks)
			s.mu.Unlock()

			return TracksUpdatedMsg{
				SourceID: playlistID,
				Tracks:   cachedEntry.Tracks,
				NextPage: cachedEntry.NextPage,
			}
		}

		log.Printf("SpotifyState: No cache found, fetching first page from API for playlist %s", playlistID)
		playlistItems, err := s.client.GetPlaylistItems(ctx, playlistID, spotify.Limit(50))
		if err != nil {
			log.Printf("SpotifyState: Error fetching playlist items: %v", err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Fetch Tracks for Playlist %s", playlistID),
				Message: err.Error(),
			}
		}

		simpleTracks := s.convertPlaylistItemsToSimpleTracks(playlistItems.Items)

		hasMore := len(playlistItems.Items) > 0 && playlistItems.Next != ""
		totalTracks := playlistItems.Total
		log.Printf("SpotifyState: Total tracks in playlist %s: %d, hasMore: %v",
			playlistID, totalTracks, hasMore)

		s.updateCacheEntry(playlistID, simpleTracks, playlistItems, hasMore, int(totalTracks), false)

		log.Printf("SpotifyState: Successfully fetched first page: %d tracks (total: %d, hasMore: %v)",
			len(simpleTracks), totalTracks, hasMore)

		return TracksUpdatedMsg{
			SourceID: playlistID,
			Tracks:   simpleTracks,
			NextPage: playlistItems,
		}
	}
}

func (s *SpotifyState) FetchNextTracksPage(ctx context.Context, sourceID spotify.ID, nextPage *spotify.PlaylistItemPage) tea.Cmd {
	return func() tea.Msg {
		if nextPage == nil || nextPage.Next == "" {
			log.Printf("SpotifyState: No next page available.")
			return nil
		}

		s.mu.Lock()
		if s.fetchingPages[nextPage.Next] {
			log.Printf("SpotifyState: Ignoring duplicate request to fetch page: %s", nextPage.Next)
			s.mu.Unlock()
			return nil
		}
		s.fetchingPages[nextPage.Next] = true
		s.mu.Unlock()

		defer func() {
			s.mu.Lock()
			delete(s.fetchingPages, nextPage.Next)
			s.mu.Unlock()
		}()

		log.Printf("SpotifyState: Fetching next page of tracks for source %s", sourceID)

		pageToFetch := *nextPage

		err := s.client.NextPage(ctx, &pageToFetch)
		if err != nil {
			if err == spotify.ErrNoMorePages {
				log.Printf("SpotifyState: No more pages available")
				s.mu.Lock()
				if entry, exists := s.tracksCache[sourceID]; exists {
					entry.HasMore = false
					entry.NextPage = nil
				}
				s.mu.Unlock()
				return nil
			}
			log.Printf("SpotifyState: Error fetching next page of tracks: %v", err)
			return ErrorMsg{
				Title:   "Failed to Fetch Next Page",
				Message: err.Error(),
			}
		}

		log.Printf("SpotifyState: Successfully fetched next page, %d new items", len(pageToFetch.Items))

		newSimpleTracks := s.convertPlaylistItemsToSimpleTracks(pageToFetch.Items)

		hasMore := len(pageToFetch.Items) > 0 && pageToFetch.Next != ""

		s.updateCacheEntry(sourceID, newSimpleTracks, &pageToFetch, hasMore, int(pageToFetch.Total), true)

		log.Printf("SpotifyState: Successfully appended %d new tracks, hasMore: %v",
			len(newSimpleTracks), hasMore)

		return TracksUpdatedMsg{
			SourceID: sourceID,
			Tracks:   newSimpleTracks,
			NextPage: &pageToFetch,
		}
	}
}

// FetchArtistTopTracks retrieves top tracks for an artist
func (s *SpotifyState) FetchArtistTopTracks(ctx context.Context, artistID spotify.ID, country string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Fetching top tracks for artist: %s", artistID)

		// Check cache first
		s.mu.RLock()
		cachedEntry, exists := s.tracksCache[artistID]
		s.mu.RUnlock()

		if exists {
			log.Printf("SpotifyState: Found cached tracks for artist %s", artistID)

			s.mu.Lock()
			s.tracks = make([]spotify.SimpleTrack, len(cachedEntry.Tracks))
			copy(s.tracks, cachedEntry.Tracks)
			s.mu.Unlock()

			return TracksUpdatedMsg{
				SourceID: artistID,
				Tracks:   cachedEntry.Tracks,
				NextPage: nil,
			}
		}

		// Fetch artist top tracks
		topTracks, err := s.client.GetArtistsTopTracks(ctx, artistID, country)
		if err != nil {
			log.Printf("SpotifyState: Error fetching artist top tracks: %v", err)
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Fetch Artist Top Tracks %s", artistID),
				Message: err.Error(),
			}
		}

		// Convert FullTrack to SimpleTrack
		simpleTracks := make([]spotify.SimpleTrack, len(topTracks))
		for i, track := range topTracks {
			simpleTracks[i] = track.SimpleTrack
		}

		// Update cache and current state
		s.updateCacheEntry(artistID, simpleTracks, nil, false, len(simpleTracks), false)

		log.Printf("SpotifyState: Successfully fetched %d artist top tracks", len(simpleTracks))

		return TracksUpdatedMsg{
			SourceID: artistID,
			Tracks:   simpleTracks,
			NextPage: nil,
		}
	}
}

// Helper function to convert playlist items to simple tracks
func (s *SpotifyState) convertPlaylistItemsToSimpleTracks(items []spotify.PlaylistItem) []spotify.SimpleTrack {
	simpleTracks := make([]spotify.SimpleTrack, 0, len(items))

	for _, item := range items {
		if item.Track.Track != nil && item.Track.Track.SimpleTrack.ID != "" {
			track := item.Track.Track.SimpleTrack

			// Fix album name issue - copy from full track if simple track album name is empty
			if track.Album.Name == "" && item.Track.Track.Album.Name != "" {
				track.Album.Name = item.Track.Track.Album.Name
			}

			simpleTracks = append(simpleTracks, track)
		}
	}

	return simpleTracks
}
