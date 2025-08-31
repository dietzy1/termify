package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DialogContent interface {
	tea.Model
	GetTitle() string
	SetSize(width, height int)
	HandleDialogKey(msg tea.KeyMsg) (bool, tea.Cmd)
	GetActions() []DialogAction
}

type DialogAction struct {
	Label string
	Key   key.Binding
	Cmd   tea.Cmd
}

type dialogModel struct {
	title       string
	message     string
	width       int
	height      int
	visible     bool
	acceptText  string
	cancelText  string
	acceptKey   key.Binding
	cancelKey   key.Binding
	focusAccept bool

	content     DialogContent
	contentMode bool
	actionFocus int
}

type DialogMsg struct {
	Accepted bool
}

type ShowDialogMsg struct {
	Title      string
	Message    string
	AcceptText string
	CancelText string
}

type ShowDialogWithContentMsg struct {
	Content DialogContent
}

type HideDialogMsg struct{}

func newDialog() dialogModel {
	return dialogModel{
		title:       "",
		message:     "",
		visible:     false,
		acceptText:  "Yes",
		cancelText:  "No",
		acceptKey:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "accept")),
		cancelKey:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		focusAccept: true,
		contentMode: false,
		actionFocus: 0,
	}
}

func (m dialogModel) Init() tea.Cmd {
	return nil
}

func (m dialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("Dialog Update called")

	switch msg := msg.(type) {
	case ShowDialogMsg:
		log.Printf("Dialog Update: ShowDialogMsg received - Title: %s", msg.Title)
		m.title = msg.Title
		m.message = msg.Message
		m.acceptText = msg.AcceptText
		m.cancelText = msg.CancelText
		m.visible = true
		m.focusAccept = true
		m.contentMode = false
		m.content = nil
		log.Printf("Dialog Update: Dialog now visible=%t", m.visible)
		return m, nil

	case ShowDialogWithContentMsg:
		log.Printf("Dialog Update: ShowDialogWithContentMsg received")
		m.content = msg.Content
		m.contentMode = true
		m.visible = true
		m.actionFocus = 0
		if m.content != nil {
			m.content.SetSize(m.width, m.height)
			if cmd := m.content.Init(); cmd != nil {
				return m, cmd
			}
		}
		return m, nil

	case HideDialogMsg:
		m.visible = false
		m.contentMode = false
		m.content = nil
		return m, nil

	case tea.KeyMsg:
		if m.contentMode && m.content != nil {
			if handled, cmd := m.content.HandleDialogKey(msg); handled {
				return m, cmd
			}

			actions := m.content.GetActions()
			if len(actions) > 0 {
				switch {
				case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "right"))):
					m.actionFocus = (m.actionFocus + 1) % len(actions)
					return m, nil
				case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab", "left"))):
					m.actionFocus = (m.actionFocus - 1 + len(actions)) % len(actions)
					return m, nil
				case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
					if m.actionFocus < len(actions) {
						m.visible = false
						m.contentMode = false
						m.content = nil
						return m, actions[m.actionFocus].Cmd
					}
				case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
					m.visible = false
					m.contentMode = false
					m.content = nil
					return m, func() tea.Msg {
						return DialogMsg{Accepted: false}
					}
				}
			}

			if updatedContent, cmd := m.content.Update(msg); updatedContent != nil {
				m.content = updatedContent.(DialogContent)
				return m, cmd
			}
		} else {
			switch {
			case key.Matches(msg, m.acceptKey):
				m.visible = false
				return m, func() tea.Msg {
					return DialogMsg{Accepted: true}
				}
			case key.Matches(msg, m.cancelKey):
				m.visible = false
				return m, func() tea.Msg {
					return DialogMsg{Accepted: false}
				}
			case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "shift+tab", "left", "right"))):
				m.focusAccept = !m.focusAccept
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.contentMode && m.content != nil {
			m.content.SetSize(m.width, m.height)
		}
		return m, nil
	}

	return m, nil
}

func (m dialogModel) View() string {
	if !m.visible {
		return ""
	}
	dialogBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2).
		Foreground(WhiteTextColor)

	if m.contentMode && m.content != nil {
		return m.renderContentDialog(dialogBoxStyle)
	} else {
		return m.renderSimpleDialog(dialogBoxStyle)
	}
}

func (m dialogModel) renderSimpleDialog(dialogBoxStyle lipgloss.Style) string {
	buttonStyle := lipgloss.NewStyle().
		Foreground(TextColor).
		Background(SecondaryColor).
		Padding(0, 3).
		MarginTop(1)

	activeButtonStyle := buttonStyle.
		Foreground(WhiteTextColor).
		Background(PrimaryColor).
		MarginRight(2).
		Underline(true)

	var acceptButton, cancelButton string
	if m.focusAccept {
		acceptButton = activeButtonStyle.Render(m.acceptText)
		cancelButton = buttonStyle.Render(m.cancelText)
	} else {
		acceptButton = buttonStyle.Render(m.acceptText)
		cancelButton = activeButtonStyle.Render(m.cancelText)
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginBottom(1)

	// Message styling
	messageStyle := lipgloss.NewStyle().
		Foreground(WhiteTextColor).
		Width(50).
		Align(lipgloss.Center)

	title := titleStyle.Render(m.title)
	message := messageStyle.Render(m.message)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, acceptButton, cancelButton)

	ui := lipgloss.JoinVertical(lipgloss.Center, title, message, buttons)

	return dialogBoxStyle.Render(ui)
}

func (m dialogModel) renderContentDialog(dialogBoxStyle lipgloss.Style) string {
	if m.content == nil {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginBottom(1)

	contentView := m.content.View()

	actions := m.content.GetActions()
	var actionButtons []string

	buttonStyle := lipgloss.NewStyle().
		Foreground(TextColor).
		Background(SecondaryColor).
		Padding(0, 3).
		MarginTop(1).
		MarginRight(1)

	activeButtonStyle := buttonStyle.
		Foreground(WhiteTextColor).
		Background(PrimaryColor).
		Underline(true)

	for i, action := range actions {
		var button string
		if i == m.actionFocus {
			button = activeButtonStyle.Render(action.Label)
		} else {
			button = buttonStyle.Render(action.Label)
		}
		actionButtons = append(actionButtons, button)
	}

	title := titleStyle.Render(m.content.GetTitle())
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, actionButtons...)

	ui := lipgloss.JoinVertical(lipgloss.Center, title, contentView, buttons)

	return dialogBoxStyle.Render(ui)
}

// IsVisible returns whether the dialog is currently visible
func (m dialogModel) IsVisible() bool {
	return m.visible
}

// Show displays the dialog with the given parameters
func (m *dialogModel) Show(title, message, acceptText, cancelText string) {
	m.title = title
	m.message = message
	m.acceptText = acceptText
	m.cancelText = cancelText
	m.visible = true
	m.focusAccept = true
}

// Hide hides the dialog
func (m *dialogModel) Hide() {
	m.visible = false
	m.contentMode = false
	m.content = nil
}

// ShowWithContent displays the dialog with injected content
func (m *dialogModel) ShowWithContent(content DialogContent) {
	m.content = content
	m.contentMode = true
	m.visible = true
	m.actionFocus = 0
	if m.content != nil {
		m.content.SetSize(m.width, m.height)
	}
}
