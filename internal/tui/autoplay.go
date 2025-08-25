package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/dietzy1/termify/internal/state"
)

func (m applicationModel) handleAutoplay() (applicationModel, tea.Cmd) {
	if !m.spotifyState.Queue.IsEmpty() {
		log.Println("Queue is not empty, playing next track from queue")
		track, err := m.spotifyState.Queue.Dequeue()
		if err != nil {
			log.Printf("Error dequeuing track from queue: %v", err)
			return m, nil
		}
		log.Printf("Playing next track from queue: %s", track.ID)
		return m, tea.Batch(
			state.UpdateQueue(),
			m.spotifyState.PlayTrack(m.ctx, track.ID),
		)
	}

	log.Println("Queue is empty, playing next track based on current view")
	switch {
	case m.focusedModel == FocusPlaylistView || m.focusedModel == FocusLibrary:
		// PlaylistView handles both normal playlists and other track views (artist, album)
		if nextTrack := m.playlistView.getNextTrack(); nextTrack != "" {
			log.Printf("Playing next track from playlist: %s", nextTrack)
			return m, m.spotifyState.PlayTrack(m.ctx, nextTrack)
		}

	case m.isSearchViewFocus():
		if nextTrack := m.searchView.GetNextTrack(m.focusedModel); nextTrack != "" {
			log.Printf("Playing next track from search: %s", nextTrack)
			return m, m.spotifyState.PlayTrack(m.ctx, nextTrack)
		}
	}
	log.Println("No next track found in current view, not playing anything")
	//TODO: Do we need to add a fallback here? // Could play from recent tracks or try to come up with a recommendation
	return m, m.spotifyState.FetchPlaybackState(m.ctx)
}
