package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dietzy1/termify/internal/state"
)

// TODO: I dont think this autoplay behaviour is needed I might have misunderstood what the fuck a spotify play context is.
// Its probaly built into the Spotify API and should be handled there by changing play context each time we change view.
// The queue however I think is fine as is to keep client side so nothing wrong there

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
