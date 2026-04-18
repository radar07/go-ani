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

// resultItem wraps a SearchResult for the list model
type resultItem struct {
	result api.SearchResult
	mode   string
}

func (r resultItem) Title() string { return r.result.Name }
func (r resultItem) Description() string {
	epCount := r.result.AvailableEpisodes.Sub
	label := "sub"
	if r.mode == "dub" {
		epCount = r.result.AvailableEpisodes.Dub
		label = "dub"
	}
	return fmt.Sprintf("%d episodes • %s", epCount, label)
}
func (r resultItem) FilterValue() string { return r.result.Name }

// resultDelegate renders list items with custom styling
type resultDelegate struct{}

func (d resultDelegate) Height() int                             { return 2 }
func (d resultDelegate) Spacing() int                            { return 0 }
func (d resultDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d resultDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ri, ok := item.(resultItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	title := ri.Title()
	desc := ri.Description()
	num := fmt.Sprintf("%d. ", index+1)

	if isSelected {
		fmt.Fprint(w, selectedStyle.Render("▸ "+num+title)+"\n") //nolint:errcheck
		fmt.Fprint(w, "    "+dimStyle.Render(desc))              //nolint:errcheck
	} else {
		fmt.Fprint(w, "  "+num+textStyle.Render(title)+"\n") //nolint:errcheck
		fmt.Fprint(w, "    "+dimStyle.Render(desc))          //nolint:errcheck
	}
}

type resultsModel struct {
	list    list.Model
	results []api.SearchResult
	mode    string
	width   int
	height  int
}

func newResultsModel(results []api.SearchResult, mode string, width, height int) resultsModel {
	items := make([]list.Item, len(results))
	for i, r := range results {
		items[i] = resultItem{result: r, mode: mode}
	}

	delegate := resultDelegate{}
	l := list.New(items, delegate, width, height-4)
	l.Title = fmt.Sprintf("📺 Results (%d found)", len(results))
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)

	return resultsModel{
		list:    l,
		results: results,
		mode:    mode,
		width:   width,
		height:  height,
	}
}

func (m resultsModel) Init() tea.Cmd {
	return nil
}

func (m resultsModel) Update(msg tea.Msg) (resultsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" && !m.list.SettingFilter() {
			if item, ok := m.list.SelectedItem().(resultItem); ok {
				_ = item // selection handled by app coordinator
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m resultsModel) View() string {
	var b strings.Builder
	b.WriteString(m.list.View())
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓: navigate • enter: select • /: filter • q: quit"))
	return b.String()
}

func (m resultsModel) selectedResult() *api.SearchResult {
	if item, ok := m.list.SelectedItem().(resultItem); ok {
		return &item.result
	}
	return nil
}
