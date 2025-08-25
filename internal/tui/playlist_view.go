package tui

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/spinner"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/dietzy1/termify/internal/state"
	"github.com/evertras/bubble-table/table"
	"github.com/zmb3/spotify/v2"
)

type clearQueuedHighlightMsg struct {
	TrackID spotify.ID
}

type TrackRowType int

const (
	TrackRowLoaded TrackRowType = iota
	TrackRowLoading
)

type TrackRow struct {
	Type     TrackRowType
	Track    *spotify.SimpleTrack
	Index    int
	IsQueued bool
}

type playlistViewModel struct {
	ctx            context.Context
	width, height  int
	table          table.Model
	isFocused      bool
	spotifyState   *state.SpotifyState
	queuedTracks   map[spotify.ID]bool // Track which songs have been queued recently
	highlightTimer *time.Timer         // Timer to clear the highlight
	spinner        spinner.Model
}

func newPlaylistView(ctx context.Context, spotifyState *state.SpotifyState) playlistViewModel {

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(PrimaryColor)
	s.Spinner.FPS = time.Second * 1 / 2

	return playlistViewModel{
		ctx:          ctx,
		table:        createPlaylistTable(),
		spotifyState: spotifyState,
		queuedTracks: make(map[spotify.ID]bool),
		spinner:      s,
	}
}

func (m playlistViewModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick)
}

// Update handles messages and updates the playlist view
func (m playlistViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		const minTableHeight = 7
		// This check is a panic safeguard
		if m.height-minTableHeight < 0 {
			m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height).WithPageSize(1)
			return m, nil
		}

		m.table = m.table.WithTargetWidth(m.width).WithMinimumHeight(m.height).WithPageSize(m.height - 6)
		log.Printf("PlaylistView width: %d, height: %d", m.width, m.height)
		return m, nil

	case spinner.TickMsg:
		var spinnerCmd, loadingSpinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		m.updateTableWithTracksAndLoading()
		return m, tea.Batch(spinnerCmd, loadingSpinnerCmd)

	case state.TracksUpdatedMsg:
		m.updateTableWithTracksAndLoading()

		if msg.NextPage != nil {
			currentPage := m.table.CurrentPage()
			pageSize := m.height - headerFooterHeight
			if pageSize <= 0 {
				pageSize = 1
			}

			totalLoadedTracks := len(m.spotifyState.GetTracks())
			maxPageWithCurrentData := (totalLoadedTracks + pageSize - 1) / pageSize

			if currentPage >= maxPageWithCurrentData-1 {
				log.Printf("PlaylistView: Prefetching next page (current: %d, max loaded: %d)",
					currentPage, maxPageWithCurrentData)
				return m, m.spotifyState.FetchNextTracksPage(m.ctx, msg.SourceID, msg.NextPage)
			}
		}

		log.Printf("PlaylistView: No prefetch needed (current page: %d, total tracks: %d)",
			m.table.CurrentPage(), len(m.spotifyState.GetTracks()))
		return m, nil

	// Handle keyboard events for table navigation
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Left, DefaultKeyMap.Right):
			var tableCmd tea.Cmd
			m.table, tableCmd = m.table.Update(msg)
			prefetchCmd := m.checkAndPrefetchIfNeeded()
			if prefetchCmd != nil {
				return m, tea.Batch(tableCmd, prefetchCmd)
			}
			return m, tableCmd

		case key.Matches(msg, DefaultKeyMap.Up, DefaultKeyMap.Down):
			var tableCmd tea.Cmd
			m.table, tableCmd = m.table.Update(msg)

			prefetchCmd := m.checkAndPrefetchIfNeeded()
			if prefetchCmd != nil {
				return m, tea.Batch(tableCmd, prefetchCmd)
			}
			return m, tableCmd

		case key.Matches(msg, DefaultKeyMap.Select):
			if track := m.getSelectedTrack(); track != nil {
				log.Printf("PlaylistView: Selected track: %s", track.ID)
				return m, m.spotifyState.PlayTrack(m.ctx, track.ID)
			}
			return m, nil

		case key.Matches(msg, DefaultKeyMap.AddToQueue):
			if track := m.getSelectedTrack(); track != nil {
				log.Printf("PlaylistView: Adding track to queue: %s", track.ID)

				m.queuedTracks[track.ID] = true
				m.updateTableWithTracksAndLoading()
				m.spotifyState.Queue.Enqueue(*track)
				return m, tea.Batch(
					state.UpdateQueue(),
					tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
						return clearQueuedHighlightMsg{TrackID: track.ID}
					}),
				)
			}
			return m, nil
		}
	case clearQueuedHighlightMsg:
		delete(m.queuedTracks, msg.TrackID)
		m.updateTableWithTracksAndLoading()
		return m, nil
	}

	// Forward all other messages to the table, but check for page changes
	oldPage := m.table.CurrentPage()
	m.table, cmd = m.table.Update(msg)
	newPage := m.table.CurrentPage()

	// If page changed, check if we need to prefetch
	if oldPage != newPage {
		prefetchCmd := m.checkAndPrefetchIfNeeded()
		if prefetchCmd != nil {
			cmd = tea.Batch(cmd, prefetchCmd)
		}
	}

	return m, cmd
}

func (m playlistViewModel) View() string {

	m.table = m.table.Border(RoundedTableBorders).
		HeaderStyle(
			lipgloss.NewStyle().
				BorderBackground(BackgroundColor).
				BorderForeground(getBorderStyle(m.isFocused)))
	m.table = m.table.WithBaseStyle(
		lipgloss.NewStyle().
			BorderBackground(BackgroundColor).
			Background(BackgroundColor).
			BorderForeground(getBorderStyle(m.isFocused)),
	)
	m.table = m.table.Focused(m.isFocused)

	selectedId := m.spotifyState.GetSelectedID()
	playlists := m.spotifyState.GetPlaylists()
	currentPage := m.table.CurrentPage()
	maxPage := m.calculateMaxPage(m.spotifyState.GetTotalTracks(selectedId))

	name := "Unknown Playlist"
	for _, playlist := range playlists {
		if playlist.ID == spotify.ID(selectedId) {
			name = playlist.Name
			break
		}
	}
	//As fallback check searchResults for the selected playlist and brute force it
	if name == "Unknown Playlist" {
		playlistSearchResults := m.spotifyState.GetSearchResultPlaylists()
		for _, playlist := range playlistSearchResults {
			if playlist.ID == spotify.ID(selectedId) {
				name = playlist.Name
				break
			}
		}

		artistsSearchResults := m.spotifyState.GetSearchResultArtists()
		for _, artist := range artistsSearchResults {
			if artist.ID == spotify.ID(selectedId) {
				name = artist.Name
				break
			}
		}
		albumsSearchResults := m.spotifyState.GetSearchResultAlbums()
		for _, album := range albumsSearchResults {
			if album.ID == spotify.ID(selectedId) {
				name = album.Name
				break
			}
		}
	}

	styledName := lipgloss.NewStyle().
		Foreground(TextColor).
		Padding(0, 1).
		Background(BackgroundColor).
		Render(name)

	styledPage := lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Padding(0, 1).
		Background(BackgroundColor).
		Render(fmt.Sprintf("| Page %d/%d", currentPage, maxPage))

	m.table = m.table.WithStaticFooter(
		styledName + styledPage,
	)

	return m.table.View()
}

func (m *playlistViewModel) SetFocus(isFocused bool) {
	m.isFocused = isFocused
}

// updateTableWithTracksAndLoading creates a combined view of loaded tracks and loading placeholders
func (m *playlistViewModel) updateTableWithTracksAndLoading() {
	selectedID := m.spotifyState.GetSelectedID()
	if selectedID == "" {
		return
	}

	loadedTracks := m.spotifyState.GetTracks()
	totalTracks := m.spotifyState.GetTotalTracks(selectedID)
	if totalTracks == 0 {
		m.table = m.table.WithRows([]table.Row{})
		return
	}

	rows := make([]table.Row, totalTracks)
	playerState := m.spotifyState.GetPlayerState()

	// Add loaded tracks
	for i, track := range loadedTracks {
		rows[i] = m.createTrackRow(track, i, &playerState)
	}

	// Add loading placeholders for remaining tracks
	for i := len(loadedTracks); i < totalTracks; i++ {
		rows[i] = m.createLoadingRow(i)
	}

	m.table = m.table.WithRows(rows)
}

func (m *playlistViewModel) createTrackRow(track spotify.SimpleTrack, index int, playerState *spotify.PlayerState) table.Row {
	artistName := "Unknown Artist"
	if len(track.Artists) > 0 {
		artistName = track.Artists[0].Name
	}

	albumName := "Unknown Album"
	if track.Album.Name != "" {
		albumName = track.Album.Name
	}

	duration := formatTrackDuration(int(track.Duration))

	title := track.Name
	if m.queuedTracks[track.ID] {
		queuedStyle := lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)
		title = title + " " + queuedStyle.Render("(Added to queue)")
	}

	var indexDisplay string
	if playerState != nil && playerState.Item != nil && string(track.ID) == string(playerState.Item.ID) {
		indexDisplay = m.spinner.View()
	} else {
		indexDisplay = fmt.Sprintf("%d", index+1)
	}

	return table.NewRow(table.RowData{
		"#":        indexDisplay,
		"title":    title,
		"artist":   artistName,
		"album":    albumName,
		"duration": duration,
	})
}

func (m *playlistViewModel) createLoadingRow(index int) table.Row {
	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true).Background(BackgroundColor)

	return table.NewRow(table.RowData{
		"#":        fmt.Sprintf("%d", index+1),
		"title":    loadingStyle.Render("Loading"),
		"artist":   "--:--",
		"album":    "--:--",
		"duration": "--:--",
	})
}

func createPlaylistTable() table.Model {
	return table.New([]table.Column{
		table.NewColumn("#", "#", 4).WithStyle(lipgloss.NewStyle().Align(lipgloss.Center)),
		table.NewFlexColumn("title", "Title", 1),   // Flex column with weight 3 4
		table.NewFlexColumn("artist", "Artist", 1), // Flex column with weight 2 4
		table.NewFlexColumn("album", "Album", 1),   // Flex column with weight 2 3
		table.NewColumn("duration", "Duration", 8).WithStyle(lipgloss.NewStyle().Align(lipgloss.Center)),
	}).WithRows([]table.Row{}).HeaderStyle(
		lipgloss.NewStyle().
			Bold(true).
			BorderForeground(BorderColor).
			Underline(true),
	).WithBaseStyle(
		lipgloss.NewStyle().
			Align(lipgloss.Left).
			BorderForeground(BorderColor).
			BorderBackground(BackgroundColor).
			Background(BackgroundColor),
	).Focused(true).HighlightStyle(
		lipgloss.NewStyle().
			Background(BackgroundColor).
			Foreground(PrimaryColor).
			Padding(0, 0, 0, 1).Bold(true),
	).Border(
		RoundedTableBorders,
	)
}

// GetNextTrack returns the ID of the next track to play when autoplay is triggered
func (m *playlistViewModel) getNextTrack() spotify.ID {
	tracks := m.spotifyState.GetTracks()
	if len(tracks) == 0 {
		log.Println("No tracks in playlist to autoplay")
		return ""
	}

	// If no track is currently playing, return the first track
	playerState := m.spotifyState.GetPlayerState()

	if playerState.Item == nil {
		log.Println("No track currently playing, returning first track")
		return spotify.ID(tracks[0].ID)
	}

	currentTrackID := playerState.Item.ID
	log.Printf("Current track ID: %s", currentTrackID)

	// Find the current track in the playlist
	for i, track := range tracks {
		if string(track.ID) == string(currentTrackID) {
			// If it's the last track, return empty (we'll use recommendations)
			if i >= len(tracks)-1 {
				log.Println("Current track is the last one in playlist")
				return ""
			}
			// Return the next track
			log.Printf("Found current track at index %d, returning next track", i)
			return spotify.ID(tracks[i+1].ID)
		}
	}

	// If current track not found in playlist, return the first track
	log.Println("Current track not found in playlist, returning first track")
	return spotify.ID(tracks[0].ID)
}

func (m *playlistViewModel) getSelectedTrack() *spotify.SimpleTrack {
	selected := m.table.HighlightedRow()
	if selected.Data == nil {
		return nil
	}

	numStr, ok := selected.Data["#"].(string)
	if !ok {
		return nil
	}

	tracks := m.spotifyState.GetTracks()
	idx, err := strconv.Atoi(numStr)
	if err != nil || idx <= 0 || idx > len(tracks) {
		// User selected a loading row, return nil
		return nil
	}

	return &tracks[idx-1]
}

// 2 for header, 4 for footer
const headerFooterHeight = 6

func (m playlistViewModel) calculateMaxPage(totalTracks int) int {
	if totalTracks == 0 {
		return 1
	}

	pageSize := m.height - headerFooterHeight
	if pageSize <= 0 {
		return 1
	}

	maxPage := (totalTracks + pageSize - 1) / pageSize
	return maxPage
}

// checkAndPrefetchIfNeeded checks if we need to prefetch more data based on current position
func (m *playlistViewModel) checkAndPrefetchIfNeeded() tea.Cmd {
	selectedID := m.spotifyState.GetSelectedID()
	if selectedID == "" {
		return nil
	}

	// Check if there are more tracks to load
	if !m.spotifyState.HasMoreTracks(selectedID) {
		return nil
	}

	// Get cache entry to check for next page
	cacheEntry, exists := m.spotifyState.GetCachedTracks(selectedID)
	if !exists || cacheEntry.NextPage == nil {
		return nil
	}

	currentPage := m.table.CurrentPage()
	pageSize := m.height - headerFooterHeight
	if pageSize <= 0 {
		pageSize = 1
	}

	totalLoadedTracks := len(cacheEntry.Tracks)
	maxPageWithCurrentData := (totalLoadedTracks + pageSize - 1) / pageSize

	// Prefetch if we're on the last page of loaded data or one page before
	if currentPage >= maxPageWithCurrentData-1 {
		log.Printf("PlaylistView: Navigation triggered prefetch (current: %d, max loaded: %d)",
			currentPage, maxPageWithCurrentData)
		return m.spotifyState.FetchNextTracksPage(m.ctx, selectedID, cacheEntry.NextPage)
	}

	return nil
}
