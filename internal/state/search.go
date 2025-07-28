package state

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type SearchResultsUpdatedMsg struct{}

func (s *SpotifyState) SearchEverything(ctx context.Context, query string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Searching for: %s", query)
		if query == "" {
			return nil
		}

		results, err := s.client.Search(ctx, query, spotify.SearchTypeTrack|spotify.SearchTypeArtist|spotify.SearchTypeAlbum|spotify.SearchTypePlaylist)
		if err != nil {
			log.Printf("SpotifyState: Error searching: %v", err)
			return ErrorMsg{
				Title:   "Search Failed",
				Message: err.Error(),
			}
		}

		log.Printf("SpotifyState: Found %d tracks", len(results.Tracks.Tracks))
		log.Printf("SpotifyState: Found %d artists", len(results.Artists.Artists))
		log.Printf("SpotifyState: Found %d albums", len(results.Albums.Albums))
		log.Printf("SpotifyState: Found %d playlists", len(results.Playlists.Playlists))

		// Filter out playlists that are null for some reason
		filteredPlaylists := make([]spotify.SimplePlaylist, 0, len(results.Playlists.Playlists))
		for _, p := range results.Playlists.Playlists {
			if p.Name != "" || p.ID != "" {
				filteredPlaylists = append(filteredPlaylists, p)
			}
		}

		s.setSearchResults(
			results.Tracks.Tracks,
			results.Artists.Artists,
			results.Albums.Albums,
			filteredPlaylists,
		)

		return SearchResultsUpdatedMsg{}
	}
}

// TODO: Work in progress
/* func (s *SpotifyState) PlayRecommendTrack( trackID spotify.ID) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Finding recommendation for track: %s", trackID)
		if trackID == "" {
			log.Printf("SpotifyState: Invalid track ID for recommendation")
			// Return ErrorMsg for invalid track ID
			return ErrorMsg{
				Title:   "Invalid Input",
				Message: "Cannot get recommendations for an invalid track ID.",
			}
		}

		// Get track information
		track, err := s.client.GetTrack(context.TODO(), trackID)
		if err != nil {
			log.Printf("SpotifyState: Error getting track info for recommendation: %v", err)
			// Return ErrorMsg if getting track fails
			return ErrorMsg{
				Title:   fmt.Sprintf("Failed to Get Track Info for %s", trackID),
				Message: err.Error(),
			}
		}

		// Extract artist name and track name to build a search query
		var artistName string
		if len(track.Artists) > 0 {
			artistName = track.Artists[0].Name
		}

		// Create search query based on artist name
		searchQuery := fmt.Sprintf("artist:%s", artistName)

		log.Printf("SpotifyState: Searching recommendations with query: %s", searchQuery)
		results, err := s.client.Search(context.TODO(), searchQuery, spotify.SearchTypeTrack, spotify.Limit(50))
		if err != nil {
			log.Printf("SpotifyState: Error searching for similar tracks: %v", err)
			// Return ErrorMsg if search fails
			return ErrorMsg{
				Title:   "Recommendation Search Failed",
				Message: fmt.Sprintf("Failed to find tracks similar to %s: %s", track.Name, err.Error()),
			}
		}

		// Filter out the original track
		var similarTracks []spotify.FullTrack
		for _, t := range results.Tracks.Tracks {
			if t.ID != trackID {
				similarTracks = append(similarTracks, t)
			}
		}

		if len(similarTracks) == 0 {
			log.Printf("SpotifyState: No different similar tracks found for %s", track.Name)
			// Return ErrorMsg if no suitable tracks are found
			return ErrorMsg{
				Title:   "No Recommendations Found",
				Message: fmt.Sprintf("Could not find any other tracks similar to %s by %s.", track.Name, artistName),
			}
		}

		// Select a random track from the similar tracks
		selectedTrack := similarTracks[rand.Intn(len(similarTracks))]
		log.Printf("SpotifyState: Playing recommended track: %s by %s", selectedTrack.Name, selectedTrack.Artists[0].Name)

		// Play the selected track (returns tea.Msg directly)
		return s.PlayTrack(ctx, selectedTrack.ID)()
	}
} */
