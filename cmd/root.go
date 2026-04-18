package cmd

import (
	"fmt"
	"strings"

	"github.com/radar07/go-ani/internal/api"
	"github.com/radar07/go-ani/internal/config"
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

func runSearch(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please provide a search query")
	}

	query := strings.Join(args, " ")
	client := api.NewClient()

	fmt.Println("🔍 Searching for \"%s\" (%s)...\n\n", query, cfg.Mode)

	results, err := client.SearchAnime(query, cfg.Mode)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	for i, r := range results {
		epCount := r.AvailableEpisodes.Sub
		if cfg.Mode == "dub" {
			epCount = r.AvailableEpisodes.Dub
		}
		fmt.Printf("  %2d. %s (%d episodes)\n", i+1, r.Name, epCount)
	}

	return nil
}

func runPlay(cmd *cobra.Command, args []string) error {
	fmt.Println("▶️  Play command - Not implemented yet")
	fmt.Printf("   Config: Player=%s, Quality=%s, Mode=%s\n",
		cfg.Player, cfg.Quality, cfg.Mode)
	if len(args) > 0 {
		fmt.Printf("   Query: %v\n", args)
	}
	if flagEpisode != "" {
		fmt.Printf("   Episode: %s\n", flagEpisode)
	}
	return nil
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
