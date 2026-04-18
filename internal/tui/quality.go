package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/radar07/go-ani/internal/api"
)

// qualityItem wraps a VideoLink for the list
type qualityItem struct {
	link api.VideoLink
}

func (q qualityItem) Title() string       { return q.link.Quality }
func (q qualityItem) Description() string { return q.link.Type }
func (q qualityItem) FilterValue() string { return q.link.Quality }

// qualityDelegate renders quality items
type qualityDelegate struct{}

func (d qualityDelegate) Height() int                             { return 1 }
func (d qualityDelegate) Spacing() int                            { return 0 }
func (d qualityDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d qualityDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	qi, ok := item.(qualityItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	label := fmt.Sprintf("%s (%s)", qi.link.Quality, qi.link.Type)

	if isSelected {
		fmt.Fprint(w, selectedStyle.Render("▸ "+label))
	} else {
		fmt.Fprint(w, "  "+textStyle.Render(label))
	}
}

type qualityModel struct {
	list   list.Model
	links  []api.VideoLink
	width  int
	height int
}

func newQualityModel(links []api.VideoLink, width, height int) qualityModel {
	items := make([]list.Item, len(links))
	for i, link := range links {
		items[i] = qualityItem{link: link}
	}

	delegate := qualityDelegate{}
	l := list.New(items, delegate, width, height-4)
	l.Title = "🎯 Select Quality"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)

	return qualityModel{
		list:   l,
		links:  links,
		width:  width,
		height: height,
	}
}

func (m qualityModel) Init() tea.Cmd {
	return nil
}

func (m qualityModel) Update(msg tea.Msg) (qualityModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m qualityModel) View() string {
	var b strings.Builder
	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓: navigate • enter: select • q: quit"))
	return b.String()
}

func (m qualityModel) selectedLink() *api.VideoLink {
	if item, ok := m.list.SelectedItem().(qualityItem); ok {
		return &item.link
	}
	return nil
}
