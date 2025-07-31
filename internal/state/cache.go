package state

import (
	"github.com/zmb3/spotify/v2"
)

type CacheEntry struct {
	Tracks      []spotify.SimpleTrack
	NextPage    *spotify.PlaylistItemPage
	HasMore     bool
	TotalTracks int
}

func (s *SpotifyState) updateCacheEntry(sourceID spotify.ID, tracks []spotify.SimpleTrack,
	nextPage *spotify.PlaylistItemPage, hasMore bool, totalTracks int, isAppend bool) {

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.tracksCache[sourceID]
	if !exists || !isAppend {
		s.tracksCache[sourceID] = &CacheEntry{
			Tracks:      tracks,
			NextPage:    nextPage,
			HasMore:     hasMore,
			TotalTracks: totalTracks,
		}
	} else {
		entry.Tracks = append(entry.Tracks, tracks...)
		entry.NextPage = nextPage
		entry.HasMore = hasMore
		entry.TotalTracks = totalTracks
	}

	if sourceID == s.selectedID {
		if isAppend && exists {
			s.tracks = append(s.tracks, tracks...)
		} else {
			s.tracks = s.tracksCache[sourceID].Tracks
		}
	}
}

func (s *SpotifyState) GetCachedTracks(sourceID spotify.ID) (*CacheEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.tracksCache[sourceID]
	if !exists {
		return nil, false
	}

	tracksCopy := make([]spotify.SimpleTrack, len(entry.Tracks))
	copy(tracksCopy, entry.Tracks)

	return &CacheEntry{
		Tracks:      tracksCopy,
		NextPage:    entry.NextPage,
		HasMore:     entry.HasMore,
		TotalTracks: entry.TotalTracks,
	}, true
}

func (s *SpotifyState) HasMoreTracks(sourceID spotify.ID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.tracksCache[sourceID]
	if !exists {
		return true
	}
	return entry.HasMore
}

func (s *SpotifyState) GetTotalTracks(sourceID spotify.ID) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.tracksCache[sourceID]
	if !exists {
		return 0
	}
	return entry.TotalTracks
}
