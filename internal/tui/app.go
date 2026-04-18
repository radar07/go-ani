package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/radar07/go-ani/internal/api"
	"github.com/radar07/go-ani/internal/config"
	"github.com/radar07/go-ani/internal/player"
)

// AppState represents the current step in the TUI flow
type AppState int

const (
	StateSearch AppState = iota
	StateResults
	StateEpisodes
	StateLoading
	StateQuality
	StatePlayback
)

// AppMode determines what the TUI does after selection
type AppMode int

const (
	ModeSearch AppMode = iota // search-only: show episodes then quit
	ModePlay                  // full play: search → select → episode → play
)

// AppOpts configures the initial state of the TUI
type AppOpts struct {
	Client  *api.Client
	Cfg     *config.Config
	Mode    AppMode
	Query   string // pre-filled search query (skip search input if set)
	Episode string // pre-selected episode (skip episode picker if set)
}

// AppModel is the top-level coordinator for all TUI states
type AppModel struct {
	// Configuration
	client *api.Client
	cfg    *config.Config
	mode   AppMode

	// Current state
	state AppState

	// Sub-models
	search   searchModel
	results  resultsModel
	episodes episodesModel
	loading  loadingModel
	quality  qualityModel
	playback playbackModel

	// Shared data across states
	selected       *api.SearchResult
	episodeList    []string
	currentEpisode string
	videoLinks     []api.VideoLink
	selectedLink   *api.VideoLink
	presetEpisode  string

	// Terminal dimensions
	width  int
	height int

	// Should we quit?
	quitting bool
}

// NewAppModel creates the TUI app with the given options
func NewAppModel(opts AppOpts) AppModel {
	m := AppModel{
		client:        opts.Client,
		cfg:           opts.Cfg,
		mode:          opts.Mode,
		search:        newSearchModel(),
		loading:       newLoadingModel("Loading..."),
		presetEpisode: opts.Episode,
		width:         80,
		height:        24,
	}

	if opts.Query != "" {
		// Skip search input — go straight to searching
		m.state = StateLoading
		m.loading = newLoadingModel(fmt.Sprintf("Searching for \"%s\"...", opts.Query))
	} else {
		m.state = StateSearch
	}

	return m
}

func (m AppModel) Init() tea.Cmd {
	var cmds []tea.Cmd

	switch m.state {
	case StateSearch:
		cmds = append(cmds, m.search.Init())
	case StateLoading:
		// We have a pre-filled query — trigger search immediately
		cmds = append(cmds, m.loading.Init())
	}

	return tea.Batch(cmds...)
}

// StartWithQuery returns an Init command that triggers an immediate search
func StartWithQuery(query string) tea.Cmd {
	return func() tea.Msg {
		return startSearchMsg{query: query}
	}
}

type startSearchMsg struct {
	query string
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			// Only quit from non-input states
			if m.state != StateSearch && !m.isFiltering() {
				m.quitting = true
				return m, tea.Quit
			}
		case "esc":
			return m.handleBack()
		}
	}

	switch m.state {
	case StateSearch:
		return m.updateSearch(msg)
	case StateResults:
		return m.updateResults(msg)
	case StateEpisodes:
		return m.updateEpisodes(msg)
	case StateLoading:
		return m.updateLoading(msg)
	case StateQuality:
		return m.updateQuality(msg)
	case StatePlayback:
		return m.updatePlayback(msg)
	}

	return m, nil
}

func (m AppModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case StateSearch:
		return m.search.View()
	case StateResults:
		return m.results.View()
	case StateEpisodes:
		return m.episodes.View()
	case StateLoading:
		return m.loading.View()
	case StateQuality:
		return m.quality.View()
	case StatePlayback:
		return m.playback.View()
	}

	return ""
}

// --- State update handlers ---

func (m AppModel) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			query := m.search.input.Value()
			if query != "" {
				m.search.loading = true
				m.search.err = nil
				return m, tea.Batch(m.search.spinner.Tick, doSearchWithMode(query, m.cfg.Mode))
			}
		}
	case SearchCompleteMsg:
		m.search.loading = false
		if msg.Err != nil {
			m.search.err = msg.Err
			return m, nil
		}
		return m.transitionToResults(msg.Results)
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	return m, cmd
}

func (m AppModel) updateResults(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startSearchMsg:
		// Triggered by StartWithQuery
		m.state = StateLoading
		m.loading = newLoadingModel(fmt.Sprintf("Searching for \"%s\"...", msg.query))
		return m, tea.Batch(m.loading.Init(), doSearchWithMode(msg.query, m.cfg.Mode))
	case SearchCompleteMsg:
		if msg.Err != nil {
			m.state = StateSearch
			m.search.err = msg.Err
			return m, nil
		}
		return m.transitionToResults(msg.Results)
	case tea.KeyMsg:
		if msg.String() == "enter" && !m.results.list.SettingFilter() {
			if result := m.results.selectedResult(); result != nil {
				m.selected = result
				m.state = StateLoading
				m.loading = newLoadingModel(fmt.Sprintf("Fetching episodes for %s...", result.Name))
				return m, tea.Batch(m.loading.Init(), doFetchEpisodes(m.client, result.ID, m.cfg.Mode))
			}
		}
	}

	var cmd tea.Cmd
	m.results, cmd = m.results.Update(msg)
	return m, cmd
}

func (m AppModel) updateEpisodes(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "enter":
			if !m.episodes.list.SettingFilter() {
				ep := m.episodes.selectedEpisode()
				if ep != "" {
					return m.startEpisodePlay(ep)
				}
			}
		case "esc":
			// Back to results (handled in handleBack)
		}
	}

	var cmd tea.Cmd
	m.episodes, cmd = m.episodes.Update(msg)
	return m, cmd
}

func (m AppModel) updateLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startSearchMsg:
		m.loading = newLoadingModel(fmt.Sprintf("Searching for \"%s\"...", msg.query))
		return m, tea.Batch(m.loading.Init(), doSearchWithMode(msg.query, m.cfg.Mode))
	case SearchCompleteMsg:
		if msg.Err != nil {
			m.state = StateSearch
			m.search = newSearchModel()
			m.search.err = msg.Err
			return m, m.search.Init()
		}
		return m.transitionToResults(msg.Results)
	case EpisodesLoadedMsg:
		if msg.Err != nil {
			m.state = StateResults
			return m, nil
		}
		m.episodeList = msg.Episodes
		if m.presetEpisode != "" {
			return m.startEpisodePlay(m.presetEpisode)
		}
		if m.mode == ModeSearch {
			// Search mode: show episodes and done
			m.state = StateEpisodes
			m.episodes = newEpisodesModel(msg.Episodes, m.selected.Name, m.width, m.height)
			return m, nil
		}
		m.state = StateEpisodes
		m.episodes = newEpisodesModel(msg.Episodes, m.selected.Name, m.width, m.height)
		return m, nil
	case SourcesLoadedMsg:
		if msg.Err != nil {
			m.state = StateEpisodes
			return m, nil
		}
		m.videoLinks = msg.Links
		// Auto-select quality if preference set or only one option
		if m.cfg.Quality != "" || len(msg.Links) == 1 {
			link, err := api.SelectQuality(msg.Links, m.cfg.Quality)
			if err != nil {
				m.state = StateEpisodes
				return m, nil
			}
			m.selectedLink = link
			return m.launchPlayer()
		}
		m.state = StateQuality
		m.quality = newQualityModel(msg.Links, m.width, m.height)
		return m, nil
	case PlayerDoneMsg:
		return m.transitionToPlayback()
	}

	var cmd tea.Cmd
	m.loading, cmd = m.loading.Update(msg)
	return m, cmd
}

func (m AppModel) updateQuality(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			if link := m.quality.selectedLink(); link != nil {
				m.selectedLink = link
				return m.launchPlayer()
			}
		}
	}

	var cmd tea.Cmd
	m.quality, cmd = m.quality.Update(msg)
	return m, cmd
}

func (m AppModel) updatePlayback(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" {
			return m.handlePlaybackAction()
		}
	}

	var cmd tea.Cmd
	m.playback, cmd = m.playback.Update(msg)
	return m, cmd
}

// --- Transition helpers ---

func (m AppModel) transitionToResults(results []api.SearchResult) (tea.Model, tea.Cmd) {
	if len(results) == 0 {
		m.state = StateSearch
		m.search = newSearchModel()
		m.search.err = fmt.Errorf("no results found")
		return m, m.search.Init()
	}

	// Auto-select if single result
	if len(results) == 1 {
		m.selected = &results[0]
		m.state = StateLoading
		m.loading = newLoadingModel(fmt.Sprintf("Fetching episodes for %s...", results[0].Name))
		return m, tea.Batch(m.loading.Init(), doFetchEpisodes(m.client, results[0].ID, m.cfg.Mode))
	}

	m.state = StateResults
	m.results = newResultsModel(results, m.cfg.Mode, m.width, m.height)
	return m, nil
}

func (m AppModel) startEpisodePlay(episode string) (tea.Model, tea.Cmd) {
	m.currentEpisode = episode

	if m.mode == ModeSearch {
		// Search mode doesn't play — just quit
		m.quitting = true
		return m, tea.Quit
	}

	m.state = StateLoading
	m.loading = newLoadingModel(fmt.Sprintf("Resolving sources for episode %s...", episode))
	return m, tea.Batch(
		m.loading.Init(),
		doFetchSources(m.client, m.selected.ID, episode, m.cfg.Mode),
	)
}

func (m AppModel) launchPlayer() (tea.Model, tea.Cmd) {
	m.state = StateLoading
	m.loading = newLoadingModel("Launching player...")

	return m, tea.Batch(
		m.loading.Init(),
		doLaunchPlayer(m.cfg, m.selected.Name, m.currentEpisode, m.selectedLink),
	)
}

func (m AppModel) transitionToPlayback() (tea.Model, tea.Cmd) {
	// Determine next/prev availability
	hasNext := false
	hasPrev := false
	for i, ep := range m.episodeList {
		if ep == m.currentEpisode {
			hasNext = i < len(m.episodeList)-1
			hasPrev = i > 0
			break
		}
	}

	m.state = StatePlayback
	m.playback = newPlaybackModel(m.selected.Name, m.currentEpisode, m.cfg.Player, hasNext, hasPrev)
	return m, nil
}

func (m AppModel) handleBack() (tea.Model, tea.Cmd) {
	switch m.state {
	case StateResults:
		m.state = StateSearch
		m.search = newSearchModel()
		return m, m.search.Init()
	case StateEpisodes:
		m.state = StateResults
		return m, nil
	case StateQuality:
		m.state = StateEpisodes
		return m, nil
	case StatePlayback:
		m.state = StateEpisodes
		return m, nil
	}
	return m, nil
}

func (m AppModel) handlePlaybackAction() (tea.Model, tea.Cmd) {
	action := m.playback.selectedAction()

	switch action {
	case actionNext:
		for i, ep := range m.episodeList {
			if ep == m.currentEpisode && i < len(m.episodeList)-1 {
				return m.startEpisodePlay(m.episodeList[i+1])
			}
		}
	case actionPrev:
		for i, ep := range m.episodeList {
			if ep == m.currentEpisode && i > 0 {
				return m.startEpisodePlay(m.episodeList[i-1])
			}
		}
	case actionReplay:
		return m.startEpisodePlay(m.currentEpisode)
	case actionPickEpisode:
		m.state = StateEpisodes
		m.episodes = newEpisodesModel(m.episodeList, m.selected.Name, m.width, m.height)
		return m, nil
	case actionQuit:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m AppModel) isFiltering() bool {
	switch m.state {
	case StateResults:
		return m.results.list.SettingFilter()
	case StateEpisodes:
		return m.episodes.list.SettingFilter()
	}
	return false
}

// --- Async commands ---

func doSearchWithMode(query, mode string) tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient()
		results, err := client.SearchAnime(query, mode)
		return SearchCompleteMsg{Results: results, Err: err}
	}
}

func doFetchEpisodes(client *api.Client, showID, mode string) tea.Cmd {
	return func() tea.Msg {
		episodes, err := client.GetEpisodes(showID, mode)
		return EpisodesLoadedMsg{Episodes: episodes, Err: err}
	}
}

func doFetchSources(client *api.Client, showID, episode, mode string) tea.Cmd {
	return func() tea.Msg {
		sources, err := client.GetEpisodeSources(showID, episode, mode)
		if err != nil {
			return SourcesLoadedMsg{Err: err}
		}
		links, err := client.GetVideoLinks(sources)
		return SourcesLoadedMsg{Links: links, Err: err}
	}
}

func doLaunchPlayer(cfg *config.Config, animeName, episode string, link *api.VideoLink) tea.Cmd {
	return func() tea.Msg {
		p, err := player.NewPlayer(cfg.Player)
		if err != nil {
			return PlayerDoneMsg{Err: err}
		}

		title := fmt.Sprintf("%s - Episode %s", animeName, episode)
		err = p.Play(player.PlayOptions{
			URL:      link.URL,
			Title:    title,
			Referrer: link.Referrer,
			NoDetach: cfg.NoDetach,
		})

		return PlayerDoneMsg{Err: err}
	}
}
