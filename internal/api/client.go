package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	apiURL      = "https://api.allanime.day/api"
	referer     = "https://allmanga.to"
	userAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0"
	baseURL     = "https://allanime.day"
	searchLimit = 40
)

// Client is the allanime API Client
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new API Client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// doGraphQL sends a POST request to the GraphQL API and returns the response body
func (c *Client) doGraphQL(query string, variables map[string]interface{}) ([]byte, error) {
	body, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// SearchAnime searches for anime by query and translation type (sub/dub)
func (c *Client) SearchAnime(query string, mode string) ([]SearchResult, error) {
	const gql = `query($search:SearchInput $limit:Int $page:Int $translationType:VaildTranslationTypeEnumType $countryOrigin:VaildCountryOriginEnumType){shows(search:$search limit:$limit page:$page translationType:$translationType countryOrigin:$countryOrigin){edges{_id name availableEpisodes __typename}}}`

	variables := map[string]interface{}{
		"search": map[string]interface{}{
			"allowAdult":   false,
			"allowUnknown": false,
			"query":        query,
		},
		"limit":           searchLimit,
		"page":            1,
		"translationType": mode,
		"countryOrigin":   "ALL",
	}

	data, err := c.doGraphQL(gql, variables)
	if err != nil {
		return nil, fmt.Errorf("search anime: %w", err)
	}

	var resp graphqlSearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	results := make([]SearchResult, 0, len(resp.Data.Shows.Edges))
	for _, show := range resp.Data.Shows.Edges {
		results = append(results, SearchResult(show))
	}
	return results, nil
}

// GetEpisodes returns sorted episode numbers for a show
func (c *Client) GetEpisodes(showID string, mode string) ([]string, error) {
	const gql = `query($showId:String!){show(_id:$showId){_id availableEpisodesDetail}}`

	variables := map[string]interface{}{
		"showId": showID,
	}

	data, err := c.doGraphQL(gql, variables)
	if err != nil {
		return nil, fmt.Errorf("get episodes: %w", err)
	}

	var resp graphqlEpisodesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse episodes response: %w", err)
	}

	var episodes []string
	switch mode {
	case "dub":
		episodes = resp.Data.Show.AvailableEpisodesDetail.Dub
	default:
		episodes = resp.Data.Show.AvailableEpisodesDetail.Sub
	}

	sort.Slice(episodes, func(i, j int) bool {
		return naturalLess(episodes[i], episodes[j])
	})

	return episodes, nil
}

// GetEpisodeSources fetches streaming sources for an episode, handling encryption
func (c *Client) GetEpisodeSources(showID, episode, mode string) ([]EpisodeSource, error) {
	const gql = `query($showId:String!,$translationType:VaildTranslationTypeEnumType!,$episodeString:String!){episode(showId:$showId translationType:$translationType episodeString:$episodeString){episodeString sourceUrls}}`

	variables := map[string]interface{}{
		"showId":          showID,
		"translationType": mode,
		"episodeString":   episode,
	}

	data, err := c.doGraphQL(gql, variables)
	if err != nil {
		return nil, fmt.Errorf("get episode sources: %w", err)
	}

	var resp graphqlEpisodeSourcesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse episode sources: %w", err)
	}

	// Handle encrypted response
	if resp.Data.Tobeparsed != "" {
		return decryptEpisodeSources(resp.Data.Tobeparsed)
	}

	// Plain response
	return resp.Data.Episode.SourceUrls, nil
}

// GetVideoLinks resolves episode sources to playable video links
func (c *Client) GetVideoLinks(sources []EpisodeSource) ([]VideoLink, error) {
	var allLinks []VideoLink

	for _, src := range sources {
		if !strings.HasPrefix(src.SourceURL, "--") {
			continue
		}

		decoded := DecodeProviderID(src.SourceURL[2:])
		if decoded == "" {
			continue
		}

		var links []VideoLink
		var err error

		if strings.HasPrefix(decoded, "http://") || strings.HasPrefix(decoded, "https://") {
			// Direct external URL
			links = append(links, VideoLink{
				URL:     decoded,
				Quality: "unknown",
				Type:    guessType(decoded),
			})
		} else {
			// Relative path — fetch from allanime.day
			links, err = c.fetchProviderLinks(baseURL + decoded)
			if err != nil {
				continue
			}
		}

		allLinks = append(allLinks, links...)
	}

	if len(allLinks) == 0 {
		return nil, fmt.Errorf("no playable links found")
	}

	// Sort by quality descending
	sort.Slice(allLinks, func(i, j int) bool {
		return qualityRank(allLinks[i].Quality) > qualityRank(allLinks[j].Quality)
	})

	return allLinks, nil
}

// SelectQuality picks the best matching link for a given quality preference
func SelectQuality(links []VideoLink, preferred string) (*VideoLink, error) {
	if len(links) == 0 {
		return nil, fmt.Errorf("no links available")
	}

	switch preferred {
	case "best", "":
		return &links[0], nil
	case "worst":
		return &links[len(links)-1], nil
	}

	// Try exact match
	for i := range links {
		if strings.Contains(links[i].Quality, preferred) {
			return &links[i], nil
		}
	}

	// Fallback to best
	return &links[0], nil
}

func guessType(url string) string {
	if strings.Contains(url, ".m3u8") {
		return "m3u8"
	}
	return "mp4"
}
