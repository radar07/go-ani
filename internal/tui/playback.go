package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type playbackAction int

const (
	actionNext playbackAction = iota
	actionPrev
	actionReplay
	actionPickEpisode
	actionQuit
)

type playbackOption struct {
	icon   string
	label  string
	action playbackAction
}

type playbackModel struct {
	options        []playbackOption
	cursor         int
	animeName      string
	currentEpisode string
	playerName     string
	hasNext        bool
	hasPrev        bool
}

func newPlaybackModel(animeName, currentEpisode, playerName string, hasNext, hasPrev bool) playbackModel {
	var options []playbackOption

	if hasNext {
		options = append(options, playbackOption{"▶", "Next episode", actionNext})
	}
	if hasPrev {
		options = append(options, playbackOption{"◀", "Previous episode", actionPrev})
	}
	options = append(options, playbackOption{"🔄", "Replay", actionReplay})
	options = append(options, playbackOption{"📋", "Pick episode", actionPickEpisode})
	options = append(options, playbackOption{"❌", "Quit", actionQuit})

	return playbackModel{
		options:        options,
		animeName:      animeName,
		currentEpisode: currentEpisode,
		playerName:     playerName,
		hasNext:        hasNext,
		hasPrev:        hasPrev,
	}
}

func (m playbackModel) Init() tea.Cmd {
	return nil
}

func (m playbackModel) Update(msg tea.Msg) (playbackModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m playbackModel) View() string {
	var b strings.Builder

	title := fmt.Sprintf("%s - Episode %s", m.animeName, m.currentEpisode)
	b.WriteString(statusStyle.Render("▶ Playing: "+title) + "\n")
	b.WriteString(dimStyle.Render("  Player: "+m.playerName) + "\n\n")
	b.WriteString(headerStyle.Render("What next?") + "\n\n")

	for i, opt := range m.options {
		label := fmt.Sprintf("%s  %s", opt.icon, opt.label)
		if i == m.cursor {
			b.WriteString(menuSelectedStyle.Render("▸ "+label) + "\n")
		} else {
			b.WriteString(menuItemStyle.Render("  "+label) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓: navigate • enter: select"))
	return b.String()
}

func (m playbackModel) selectedAction() playbackAction {
	if m.cursor >= 0 && m.cursor < len(m.options) {
		return m.options[m.cursor].action
	}
	return actionQuit
}
