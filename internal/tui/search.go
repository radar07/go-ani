package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type searchModel struct {
	input   textinput.Model
	spinner spinner.Model
	loading bool
	err     error
}

func newSearchModel() searchModel {
	ti := textinput.New()
	ti.Placeholder = "Search anime..."
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 40

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return searchModel{
		input:   ti,
		spinner: sp,
	}
}

func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Enter and SearchComplete handled by app coordinator
		m.err = nil
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	case SearchCompleteMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
	case ErrMsg:
		m.loading = false
		m.err = msg.Err
	}

	if !m.loading {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m searchModel) View() string {
	var s string

	s += headerStyle.Render("🔍 Search Anime") + "\n\n"
	s += m.input.View() + "\n"

	if m.loading {
		s += "\n" + m.spinner.View() + dimStyle.Render(" Searching...") + "\n"
	}

	if m.err != nil {
		s += "\n" + errorStyle.Render("Error: "+m.err.Error()) + "\n"
	}

	s += helpStyle.Render("enter: search • ctrl+c: quit")
	return s
}
