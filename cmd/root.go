package cmd

import (
	"fmt"
	"strings"

	"github.com/radar07/go-ani/internal/api"
	"github.com/radar07/go-ani/internal/config"
	"github.com/radar07/go-ani/internal/player"
	"github.com/radar07/go-ani/internal/prompt"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var (
	cfg *config.Config

	// Global flags
	flagPlayer  string
	flagQuality string
	flagDub     bool

	// Command-specific flags
	flagEpisode string
)

var rootCmd = &cobra.Command{
	Use:   "go-ani",
	Short: "Watch anime from the terminal",
	Long: `go-ani is a CLI tool to search, stream, and download anime.

Built with Go, inspired by ani-cli.`,
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load config
		cfg = config.LoadConfig()

		// Override with flags if set
		if flagPlayer != "" {
			cfg.Player = flagPlayer
		}
		if flagQuality != "" {
			cfg.Quality = flagQuality
		}
		if flagDub {
			cfg.Mode = "dub"
		}
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for anime",
	Long:  "Search for anime by title and display results",
	Args:  cobra.MinimumNArgs(0),
	RunE:  runSearch,
}

var playCmd = &cobra.Command{
	Use:   "play [query]",
	Short: "Play anime episodes",
	Long:  "Search for anime and play episodes",
	Args:  cobra.MinimumNArgs(0),
	RunE:  runPlay,
}

var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue watching from history",
	Long:  "Resume watching anime from where you left off",
	RunE:  runContinue,
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show watch history",
	Long:  "Display list of watched anime",
	RunE:  runHistory,
}

var downloadCmd = &cobra.Command{
	Use:   "download [query]",
	Short: "Download anime episodes",
	Long:  "Search for anime and download episodes",
	Args:  cobra.MinimumNArgs(0),
	RunE:  runDownload,
}

// searchAndSelect searches for anime and lets the user pick one.
// Returns the selected result and its episode list.
func searchAndSelect(client *api.Client, query string) (*api.SearchResult, []string, error) {
	fmt.Printf("🔍 Searching for \"%s\" (%s)...\n\n", query, cfg.Mode)

	results, err := client.SearchAnime(query, cfg.Mode)
	if err != nil {
		return nil, nil, fmt.Errorf("search failed: %w", err)
	}
	if len(results) == 0 {
		return nil, nil, fmt.Errorf("no results found")
	}

	// Build display items
	items := make([]string, len(results))
	for i, r := range results {
		epCount := r.AvailableEpisodes.Sub
		if cfg.Mode == "dub" {
			epCount = r.AvailableEpisodes.Dub
		}
		items[i] = fmt.Sprintf("%s (%d episodes)", r.Name, epCount)
	}

	idx, err := prompt.SelectFromList("Select anime", items)
	if err != nil {
		return nil, nil, err
	}

	selected := &results[idx]
	fmt.Printf("\n📺 Selected: %s\n", selected.Name)

	// Fetch episodes
	fmt.Println("   Fetching episodes...")
	episodes, err := client.GetEpisodes(selected.ID, cfg.Mode)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch episodes: %w", err)
	}
	if len(episodes) == 0 {
		return nil, nil, fmt.Errorf("no episodes available")
	}

	return selected, episodes, nil
}

func runSearch(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide a search query")
	}

	query := strings.Join(args, " ")
	client := api.NewClient()

	selected, episodes, err := searchAndSelect(client, query)
	if err != nil {
		return err
	}

	fmt.Printf("\n📋 Episodes for %s:\n", selected.Name)
	// Display episodes in columns
	for i, ep := range episodes {
		fmt.Printf("  %4s", ep)
		if (i+1)%10 == 0 {
			fmt.Println()
		}
	}
	if len(episodes)%10 != 0 {
		fmt.Println()
	}

	return nil
}

func runPlay(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide a search query")
	}

	query := strings.Join(args, " ")
	client := api.NewClient()

	selected, episodes, err := searchAndSelect(client, query)
	if err != nil {
		return err
	}

	// Determine episode
	epNo := flagEpisode
	if epNo == "" {
		// Show episode list and prompt
		fmt.Printf("\n📋 Episodes (%d total):\n", len(episodes))
		for i, ep := range episodes {
			fmt.Printf("  %4s", ep)
			if (i+1)%10 == 0 {
				fmt.Println()
			}
		}
		if len(episodes)%10 != 0 {
			fmt.Println()
		}

		input, err := prompt.ReadInput(fmt.Sprintf("\nEpisode number [1-%s]: ", episodes[len(episodes)-1]))
		if err != nil {
			return err
		}
		epNo = strings.TrimSpace(input)
	}

	// Validate episode exists
	found := false
	for _, ep := range episodes {
		if ep == epNo {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("episode %s not found", epNo)
	}

	// Fetch sources
	fmt.Printf("\n🔗 Fetching sources for episode %s...\n", epNo)
	sources, err := client.GetEpisodeSources(selected.ID, epNo, cfg.Mode)
	if err != nil {
		return fmt.Errorf("fetch sources: %w", err)
	}

	// Resolve to playable links
	links, err := client.GetVideoLinks(sources)
	if err != nil {
		return fmt.Errorf("resolve links: %w", err)
	}

	// Select quality
	link, err := api.SelectQuality(links, cfg.Quality)
	if err != nil {
		return fmt.Errorf("select quality: %w", err)
	}

	fmt.Printf("   Quality: %s | Type: %s\n", link.Quality, link.Type)

	// Launch player
	p, err := player.NewPlayer(cfg.Player)
	if err != nil {
		return fmt.Errorf("init player: %w", err)
	}

	title := fmt.Sprintf("%s - Episode %s", selected.Name, epNo)
	fmt.Printf("\n▶️  Playing: %s [%s]\n", title, p.Name())

	return p.Play(player.PlayOptions{
		URL:      link.URL,
		Title:    title,
		Referrer: link.Referrer,
		NoDetach: cfg.NoDetach,
	})
}

func runContinue(cmd *cobra.Command, args []string) error {
	fmt.Println("⏭️  Continue command - Not implemented yet")
	fmt.Printf("   History path: %s\n", cfg.HistoryPath)
	return nil
}

func runHistory(cmd *cobra.Command, args []string) error {
	fmt.Println("📜 History command - Not implemented yet")
	fmt.Printf("   History path: %s\n", cfg.HistoryPath)
	return nil
}

func runDownload(cmd *cobra.Command, args []string) error {
	fmt.Println("⬇️  Download command - Not implemented yet")
	fmt.Printf("   Download dir: %s\n", cfg.DownloadDir)
	if len(args) > 0 {
		fmt.Printf("   Query: %v\n", args)
	}
	if flagEpisode != "" {
		fmt.Printf("   Episode: %s\n", flagEpisode)
	}
	return nil
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags (available to all commands)
	rootCmd.PersistentFlags().StringVarP(&flagPlayer, "player", "p", "",
		"Video player to use (mpv/vlc)")
	rootCmd.PersistentFlags().StringVarP(&flagQuality, "quality", "q", "",
		"Video quality (best/1080p/720p/worst)")
	rootCmd.PersistentFlags().BoolVar(&flagDub, "dub", false,
		"Use dubbed version (default: sub)")

	// Play command flags
	playCmd.Flags().StringVarP(&flagEpisode, "episode", "e", "",
		"Episode number or range (e.g., '1' or '1-5')")

	// Download command flags
	downloadCmd.Flags().StringVarP(&flagEpisode, "episode", "e", "",
		"Episode number or range (e.g., '1' or '1-5')")

	// Add commands to root
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(playCmd)
	rootCmd.AddCommand(continueCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(downloadCmd)
}
