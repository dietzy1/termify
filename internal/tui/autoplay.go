package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func (m applicationModel) handleAutoplay() (applicationModel, tea.Cmd) {

	queue := m.spotifyState.GetQueue()

	log.Println("Autoplaying next track", queue)
	for _, track := range queue {
		log.Println("Queue contains track:", track.Name)
	}

	if m.spotifyState.IsQueueEmpty() {
		log.Println("Queue is empty, playing next track based on current view")
		// Play the next track based on the current view
		switch {
		case m.focusedModel == FocusPlaylistView || m.focusedModel == FocusLibrary:
			// PlaylistView handles both normal playlists and other track views (artist, album)
			if nextTrack := m.playlistView.getNextTrack(); nextTrack != "" {
				log.Printf("Playing next track from playlist: %s", nextTrack)
				return m, m.spotifyState.PlayTrack(nextTrack)
			}

		case m.isSearchViewFocus():
			if nextTrack := m.searchView.GetNextTrack(m.focusedModel); nextTrack != "" {
				log.Printf("Playing next track from search: %s", nextTrack)
				return m, m.spotifyState.PlayTrack(nextTrack)
			}
		}
		//TODO: Do we need to add a fallback here?
		log.Println("No next track found in current view, not playing anything")
		// Could play from recent tracks or try to come up with a recommendation
		return m, nil
	}

	// Just refresh playback state if queue is not empty
	return m, m.spotifyState.FetchPlaybackState()
}
