package state

import (
	"context"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zmb3/spotify/v2"
)

type SearchResultsUpdatedMsg struct {
	Err error
}

func (s *SpotifyState) SearchEverything(query string) tea.Cmd {
	return func() tea.Msg {
		log.Printf("SpotifyState: Searching for: %s", query)
		if query == "" {
			log.Printf("SpotifyState: Invalid query")
			return SearchResultsUpdatedMsg{
				Err: fmt.Errorf("invalid query"),
			}
		}

		results, err := s.client.Search(context.TODO(), query, spotify.SearchTypeTrack|spotify.SearchTypeArtist|spotify.SearchTypeAlbum|spotify.SearchTypePlaylist)
		if err != nil {
			log.Printf("SpotifyState: Error searching for tracks: %v", err)
			return SearchResultsUpdatedMsg{
				Err: err,
			}
		}

		log.Printf("SpotifyState: Found %d tracks", len(results.Tracks.Tracks))
		log.Printf("SpotifyState: Found %d artists", len(results.Artists.Artists))
		log.Printf("SpotifyState: Found %d albums", len(results.Albums.Albums))
		log.Printf("SpotifyState: Found %d playlists", len(results.Playlists.Playlists))

		// Filter out playlists that are null for some reason
		for i := 0; i < len(results.Playlists.Playlists); i++ {
			if results.Playlists.Playlists[i].Name == "" && results.Playlists.Playlists[i].ID == "" {
				results.Playlists.Playlists = append(results.Playlists.Playlists[:i], results.Playlists.Playlists[i+1:]...)
				i--
			}
		}

		s.setSearchResults(
			results.Tracks.Tracks,
			results.Artists.Artists,
			results.Albums.Albums,
			results.Playlists.Playlists,
		)

		return SearchResultsUpdatedMsg{
			Err: nil,
		}
	}
}

// TODO: Work in progress
func (s *SpotifyState) PlayRecommendTrack(trackID spotify.ID) tea.Cmd {

	// Use the ID to perform a search similar to performing a recommendation so use the details from the track to recommend a new one
	// We must use the client search function to do it

	log.Printf("SpotifyState: Finding recommendation for track: %s", trackID)
	if trackID == "" {
		log.Printf("SpotifyState: Invalid track ID")
	}

	// Get track information
	track, err := s.client.GetTrack(context.TODO(), trackID)
	if err != nil {
		log.Printf("SpotifyState: Error getting track info: %v", err)
	}

	// Extract artist name and track name to build a search query
	var artistName string
	if len(track.Artists) > 0 {
		artistName = track.Artists[0].Name
	}

	// Create search query based on artist name or genre
	searchQuery := artistName
	if len(track.Album.Name) > 0 {
		// Optionally include album info for better matching
		searchQuery = fmt.Sprintf("artist:%s", artistName)
	}

	log.Printf("SpotifyState: Searching with query: %s", searchQuery)
	results, err := s.client.Search(context.TODO(), searchQuery, spotify.SearchTypeTrack, spotify.Limit(50))
	if err != nil {
		log.Printf("SpotifyState: Error searching for similar tracks: %v", err)

	}

	// Check if we found any tracks
	if len(results.Tracks.Tracks) == 0 {
		log.Printf("SpotifyState: No similar tracks found")
	}

	log.Printf("SpotifyState: Found %d similar tracks", len(results.Tracks.Tracks))

	// Filter out the original track and pick a random one from the results
	var similarTracks []spotify.FullTrack
	for _, t := range results.Tracks.Tracks {
		if t.ID != trackID {
			similarTracks = append(similarTracks, t)
		}
	}

	if len(similarTracks) == 0 {
		log.Printf("SpotifyState: No different tracks found")

	}

	// Select a track - for simplicity using the first different track found
	// but you could randomize this or use some other selection criteria
	selectedTrack := similarTracks[0]
	log.Printf("SpotifyState: Playing recommended track: %s by %s", selectedTrack.Name, selectedTrack.Artists[0].Name)
	// Play the selected track
	return s.PlayTrack(selectedTrack.ID)
}
