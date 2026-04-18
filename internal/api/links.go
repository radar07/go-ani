package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// providerResponse represents the JSON returned by allanime.day embed endpoints
type providerResponse struct {
	Links []struct {
		Link          string `json:"link"`
		MP4           bool   `json:"mp4"`
		ResolutionStr string `json:"resolutionStr"`
		HLS           bool   `json:"hls"`
		Src           string `json:"src"`
	} `json:"links"`
}

// fetchProviderLinks fetches and parses video links from a provider embed URL
func (c *Client) fetchProviderLinks(url string) ([]VideoLink, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Referer", referer)
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch provider: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read provider response: %w", err)
	}

	var pResp providerResponse
	if err := json.Unmarshal(body, &pResp); err != nil {
		return nil, fmt.Errorf("parse provider response: %w", err)
	}

	var links []VideoLink
	for _, l := range pResp.Links {
		if l.Link == "" {
			continue
		}

		link := VideoLink{
			URL:     l.Link,
			Quality: l.ResolutionStr,
		}

		if l.HLS || strings.Contains(l.Link, ".m3u8") {
			link.Type = "m3u8"
			links = append(links, parseMasterPlaylist(c, link)...)
		} else {
			link.Type = "mp4"
			if link.Quality == "" {
				link.Quality = "Mp4"
			}
			links = append(links, link)
		}
	}

	return links, nil
}

// parseMasterPlaylist fetches an M3U8 master playlist and extracts variant streams
func parseMasterPlaylist(c *Client, master VideoLink) []VideoLink {
	req, err := http.NewRequest("GET", master.URL, nil)
	if err != nil {
		return []VideoLink{master}
	}
	if master.Referrer != "" {
		req.Header.Set("Referer", master.Referrer)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return []VideoLink{master}
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []VideoLink{master}
	}

	content := string(body)
	if !strings.Contains(content, "#EXTM3U") {
		return []VideoLink{master}
	}

	// If it doesn't contain #EXT-X-STREAM-INF, it's not a master playlist
	if !strings.Contains(content, "#EXT-X-STREAM-INF") {
		return []VideoLink{master}
	}

	baseURL := master.URL[:strings.LastIndex(master.URL, "/")+1]
	lines := strings.Split(content, "\n")
	var links []VideoLink

	for i, line := range lines {
		if !strings.HasPrefix(line, "#EXT-X-STREAM-INF") {
			continue
		}
		if i+1 >= len(lines) {
			continue
		}

		streamURL := strings.TrimSpace(lines[i+1])
		if streamURL == "" || strings.HasPrefix(streamURL, "#") {
			continue
		}

		// Make relative URLs absolute
		if !strings.HasPrefix(streamURL, "http") {
			streamURL = baseURL + streamURL
		}

		quality := extractResolution(line)
		links = append(links, VideoLink{
			URL:      streamURL,
			Quality:  quality,
			Type:     "m3u8",
			Referrer: master.Referrer,
		})
	}

	if len(links) == 0 {
		return []VideoLink{master}
	}
	return links
}

// extractResolution parses resolution from an EXT-X-STREAM-INF line
func extractResolution(line string) string {
	// Look for RESOLUTION=WxH
	if idx := strings.Index(line, "RESOLUTION="); idx >= 0 {
		rest := line[idx+11:]
		if end := strings.IndexAny(rest, ",\n\r "); end > 0 {
			rest = rest[:end]
		}
		// Extract height from WxH
		if parts := strings.Split(rest, "x"); len(parts) == 2 {
			return parts[1] + "p"
		}
	}
	return "unknown"
}

// qualityRank returns a numeric ranking for quality comparison (higher is better)
func qualityRank(quality string) int {
	q := strings.TrimSuffix(strings.ToLower(quality), "p")
	if n, err := strconv.Atoi(q); err == nil {
		return n
	}
	// Non-numeric qualities get low rank
	switch strings.ToLower(quality) {
	case "mp4":
		return 500
	default:
		return 0
	}
}

// naturalLess compares episode number strings numerically when possible
func naturalLess(a, b string) bool {
	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	if errA == nil && errB == nil {
		return fa < fb
	}
	return a < b
}
