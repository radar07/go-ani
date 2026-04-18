package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// episodeItem wraps an episode number string for the list
type episodeItem struct {
	number string
}

func (e episodeItem) Title() string       { return "Episode " + e.number }
func (e episodeItem) Description() string { return "" }
func (e episodeItem) FilterValue() string { return e.number }

// episodeDelegate renders episode items
type episodeDelegate struct{}

func (d episodeDelegate) Height() int                             { return 1 }
func (d episodeDelegate) Spacing() int                            { return 0 }
func (d episodeDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d episodeDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ep, ok := item.(episodeItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	label := fmt.Sprintf("Episode %s", ep.number)

	if isSelected {
		fmt.Fprint(w, selectedStyle.Render("▸ "+label)) //nolint:errcheck
	} else {
		fmt.Fprint(w, "  "+textStyle.Render(label)) //nolint:errcheck
	}
}

type episodesModel struct {
	list      list.Model
	episodes  []string
	animeName string
	width     int
	height    int
}

func newEpisodesModel(episodes []string, animeName string, width, height int) episodesModel {
	items := make([]list.Item, len(episodes))
	for i, ep := range episodes {
		items[i] = episodeItem{number: ep}
	}

	delegate := episodeDelegate{}
	l := list.New(items, delegate, width, height-4)
	l.Title = fmt.Sprintf("📋 %s (%d episodes)", animeName, len(episodes))
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)

	return episodesModel{
		list:      l,
		episodes:  episodes,
		animeName: animeName,
		width:     width,
		height:    height,
	}
}

func (m episodesModel) Init() tea.Cmd {
	return nil
}

func (m episodesModel) Update(msg tea.Msg) (episodesModel, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m episodesModel) View() string {
	var b strings.Builder
	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓: navigate • enter: play • /: filter • esc: back • q: quit"))
	return b.String()
}

func (m episodesModel) selectedEpisode() string {
	if item, ok := m.list.SelectedItem().(episodeItem); ok {
		return item.number
	}
	return ""
}
