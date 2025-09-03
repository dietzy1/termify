package tui

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dietzy1/termify/internal/state"
)

var _ tea.Model = (*queueModel)(nil)

type queueItem struct {
	title string
	desc  string
	uri   string
}

func (p queueItem) Title() string       { return p.title }
func (p queueItem) Description() string { return p.desc }
func (p queueItem) FilterValue() string { return p.title }

type queueModel struct {
	height       int
	list         list.Model
	spotifyState *state.SpotifyState
	isFocused    bool
}

type numberedDelegate struct {
	isFocused bool
}

func (d numberedDelegate) Height() int                             { return 2 }
func (d numberedDelegate) Spacing() int                            { return 1 }
func (d numberedDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d numberedDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(queueItem)
	if !ok {
		return
	}

	const itemWidth = 28
	isSelected := index == m.Index()

	titleColor := WhiteTextColor
	descColor := TextColor
	borderColor := WhiteTextColor

	if isSelected && d.isFocused {
		titleColor = PrimaryColor
		descColor = PrimaryColor
		borderColor = PrimaryColor
	} else if isSelected {
		borderColor = WhiteTextColor
	}

	var titleStyle, descStyle lipgloss.Style

	if isSelected {
		titleStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(borderColor).
			Foreground(titleColor).
			Padding(0, 0, 0, 1).
			Bold(true).
			Width(itemWidth).
			MaxWidth(itemWidth)

		descStyle = lipgloss.NewStyle().
			Foreground(descColor).
			Padding(0, 0, 0, 2).
			Width(itemWidth).
			MaxWidth(itemWidth)
	} else {
		titleStyle = lipgloss.NewStyle().
			Foreground(titleColor).
			Padding(0, 0, 0, 2).
			Width(itemWidth).
			MaxWidth(itemWidth)

		descStyle = lipgloss.NewStyle().
			Foreground(descColor).
			Padding(0, 0, 0, 2).
			Width(itemWidth).
			MaxWidth(itemWidth)
	}

	numberedTitle := fmt.Sprintf("%d. %s", index+1, i.title)

	numberPrefix := fmt.Sprintf("%d. ", index)
	descIndent := len(numberPrefix) // +1 for extra space

	if isSelected {
		descStyle = lipgloss.NewStyle().
			Foreground(descColor).
			Padding(0, 0, 0, 2+descIndent).
			Width(itemWidth).
			MaxWidth(itemWidth)
	} else {
		descStyle = lipgloss.NewStyle().
			Foreground(descColor).
			Padding(0, 0, 0, 2+descIndent).
			Width(itemWidth).
			MaxWidth(itemWidth)
	}

	fmt.Fprint(w, titleStyle.Render(numberedTitle))
	fmt.Fprint(w, "\n")
	fmt.Fprint(w, descStyle.Render(i.desc))
}

func newQueue(spotifyState *state.SpotifyState) queueModel {
	delegate := numberedDelegate{isFocused: false}

	const itemWidth = 28
	l := list.New([]list.Item{}, delegate, itemWidth+2, 0)
	l.Title = "Next in Queue"
	l.Styles.TitleBar = lipgloss.NewStyle().
		Padding(0, 0, 1, 2).
		Width(itemWidth + 2)
	l.Styles.Title = lipgloss.NewStyle().
		Background(BorderColor).
		Foreground(WhiteTextColor).
		Padding(0, 1)

	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(TextColor).
		Padding(0, 2).
		Width(itemWidth + 2).
		MaxWidth(itemWidth + 2)

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return queueModel{
		list:         l,
		spotifyState: spotifyState,
		isFocused:    false,
	}
}

func (m queueModel) Init() tea.Cmd {
	return nil
}

func (m queueModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.list.SetHeight(m.height - 3)
		return m, nil

	case state.QueueUpdatedMsg:
		m.list.SetItems(m.convertQueueToItems())

	case tea.KeyMsg:
		switch {
		// Delete key to clear the specific item
		case key.Matches(msg, DefaultKeyMap.Copy):
			_, err := m.spotifyState.Queue.PopAt(m.list.Index())
			if err != nil {
				log.Println("Error popping track from queue:", err)
			}

			return m, state.UpdateQueue()

		// Handle select key to play the selected track
		case key.Matches(msg, DefaultKeyMap.Select):

			track, err := m.spotifyState.Queue.PopAt(m.list.Index())
			if err != nil {
				log.Println("Error popping track from queue:", err)
			}

			return m, tea.Batch(
				m.spotifyState.PlayTrack(context.TODO(), track.ID),
				state.UpdateQueue(),
			)

		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m queueModel) View() string {
	delegate := numberedDelegate{isFocused: m.isFocused}
	m.list.SetDelegate(delegate)

	if len(m.list.Items()) == 0 {
		titleStyle := lipgloss.NewStyle().
			Background(BorderColor).
			Foreground(WhiteTextColor).
			Padding(0, 1)

		titleBarStyle := lipgloss.NewStyle().
			Padding(0, 0, 1, 2).
			Width(30)

		title := titleBarStyle.Render(titleStyle.Render("Next in Queue"))

		emptyStyle := lipgloss.NewStyle().
			Width(28).
			Padding(0, 0, 0, 2).
			Foreground(TextColor).
			Italic(true)

		emptyMessage := emptyStyle.Render("No tracks in queue\nAdd some music to get started!")

		content := lipgloss.JoinVertical(lipgloss.Left, title, emptyMessage)

		return lipgloss.NewStyle().
			Height(m.height - 2).
			MaxHeight(m.height).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(getBorderStyle(m.isFocused)).
			Render(content)
	}

	return lipgloss.NewStyle().
		Height(m.height - 2).
		MaxHeight(m.height).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(getBorderStyle(m.isFocused)).
		Render(m.list.View())
}

func (m queueModel) convertQueueToItems() []list.Item {
	queueItems := m.spotifyState.Queue.List()

	log.Println("Queue items found in queue model: ", queueItems)

	items := make([]list.Item, 0, len(queueItems))
	for _, item := range queueItems {
		title := item.Name
		if title == "" {
			title = "Untitled Playlist"
		}
		desc := "Unknown Artist"
		if len(item.Artists) > 0 {
			desc = item.Artists[0].Name
		}
		if desc == "" {
			desc = "Unknown Artist"
		}

		items = append(items, queueItem{
			title: title,
			desc:  desc,
			uri:   string(item.URI),
		})
	}
	return items
}
