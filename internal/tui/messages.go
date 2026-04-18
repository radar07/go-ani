package tui

import (
	"github.com/radar07/go-ani/internal/api"
)

// SearchCompleteMsg is sent when a search API call finishes
type SearchCompleteMsg struct {
	Results []api.SearchResult
	Err     error
}

// EpisodesLoadedMsg is sent when the episode list is fetched
type EpisodesLoadedMsg struct {
	Episodes []string
	Err      error
}

// SourcesLoadedMsg is sent when video sources are resolved to playable links
type SourcesLoadedMsg struct {
	Links []api.VideoLink
	Err   error
}

// PlayerDoneMsg is sent when the player exits
type PlayerDoneMsg struct {
	Err error
}

// ErrMsg wraps a generic error
type ErrMsg struct {
	Err error
}

// BackToSearchMsg signals a return to the search input
type BackToSearchMsg struct{}

// BackToEpisodesMsg signals a return to the episode selector
type BackToEpisodesMsg struct{}

// PlayEpisodeMsg signals that a specific episode should be played
type PlayEpisodeMsg struct {
	Episode string
}

// QuitMsg signals the app should exit
type QuitMsg struct{}
