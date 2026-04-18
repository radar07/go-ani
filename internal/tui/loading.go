package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type loadingModel struct {
	spinner spinner.Model
	message string
}

func newLoadingModel(message string) loadingModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return loadingModel{
		spinner: sp,
		message: message,
	}
}

func (m loadingModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m loadingModel) Update(msg tea.Msg) (loadingModel, tea.Cmd) {
	switch msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *loadingModel) SetMessage(message string) {
	m.message = message
}

func (m loadingModel) View() string {
	return "\n" + m.spinner.View() + dimStyle.Render(" "+m.message) + "\n"
}
