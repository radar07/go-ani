package api

// SearchResult represents an anime found via search
type SearchResult struct {
	ID                string
	Name              string
	AvailableEpisodes AvailableEpisodes
}

// AvailableEpisodes holds episode counts per translation type
type AvailableEpisodes struct {
	Sub int `json:"sub"`
	Dub int `json:"dub"`
	Raw int `json:"raw"`
}

// EpisodeSource represents a streaming source for an episode
type EpisodeSource struct {
	SourceURL  string  `json:"sourceUrl"`
	SourceName string  `json:"sourceName"`
	Priority   float64 `json:"priority"`
	Type       string  `json:"type"`
}

// VideoLink represents a resolved playable video URL
type VideoLink struct {
	URL         string
	Quality     string // e.g. "1080p", "720p", "Mp4"
	Type        string // "m3u8" or "mp4"
	Referrer    string
	SubtitleURL string
}

// --- Internal GraphQL response types ---

type graphqlSearchResponse struct {
	Data struct {
		Shows struct {
			Edges []graphqlShow `json:"edges"`
		} `json:"shows"`
	} `json:"data"`
}

type graphqlShow struct {
	ID                string            `json:"_id"`
	Name              string            `json:"name"`
	AvailableEpisodes AvailableEpisodes `json:"availableEpisodes"`
}

type graphqlEpisodesResponse struct {
	Data struct {
		Show struct {
			ID                      string `json:"_id"`
			AvailableEpisodesDetail struct {
				Sub []string `json:"sub"`
				Dub []string `json:"dub"`
				Raw []string `json:"raw"`
			} `json:"availableEpisodesDetail"`
		} `json:"show"`
	} `json:"data"`
}

type graphqlEpisodeSourcesResponse struct {
	Data struct {
		M          string `json:"_m"`
		Tobeparsed string `json:"tobeparsed"`
		Episode    struct {
			EpisodeString string          `json:"episodeString"`
			SourceUrls    []EpisodeSource `json:"sourceUrls"`
		} `json:"episode"`
	} `json:"data"`
}

type decryptedEpisodeResponse struct {
	Episode struct {
		EpisodeString string          `json:"episodeString"`
		SourceUrls    []EpisodeSource `json:"sourceUrls"`
	} `json:"episode"`
}
