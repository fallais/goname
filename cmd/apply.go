package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"goname/internal/models"
	"goname/pkg/services"
)

var (
	applyInputDir    string
	applyRecursive   bool
	applyDryRun      bool
	applyTmdbAPIKey  string
	applyMediaType   string
	applyInteractive bool
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the planned renames to video files using TMDB data",
	Long: `Apply renames video files in a directory by fetching information from The Movie Database (TMDB).

This command is similar to 'terraform apply' - it performs the actual renaming of files
based on TMDB lookups. Use 'goname plan' first to preview what changes will be made.

Examples:
  # Apply renames for movies in current directory
  goname apply --type movie --api-key YOUR_API_KEY
  
  # Apply renames for TV shows recursively with dry-run
  goname apply --dir /path/to/shows --type tv --recursive --dry-run --api-key YOUR_API_KEY
  
  # Interactive mode for manual confirmation
  goname apply --interactive --api-key YOUR_API_KEY`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := runApply(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	applyCmd.Flags().StringVarP(&applyInputDir, "dir", "d", ".", "Directory to scan for video files")
	applyCmd.Flags().BoolVarP(&applyRecursive, "recursive", "r", false, "Scan directories recursively")
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "Show what would be renamed without actually renaming")
	applyCmd.Flags().StringVar(&applyTmdbAPIKey, "api-key", "", "TMDB API key (can also be set via TMDB_API_KEY env var)")
	applyCmd.Flags().StringVarP(&applyMediaType, "type", "t", "auto", "Media type: movie, tv, or auto")
	applyCmd.Flags().BoolVarP(&applyInteractive, "interactive", "i", false, "Interactive mode for manual confirmation")

	// Bind flags to viper
	viper.BindPFlag("tmdb.api_key", applyCmd.Flags().Lookup("api-key"))
	viper.BindEnv("tmdb.api_key", "TMDB_API_KEY")
}

func runApply() error {
	// Get API key from flag, config, or environment
	apiKey := applyTmdbAPIKey
	if apiKey == "" {
		apiKey = viper.GetString("tmdb.api_key")
	}
	if apiKey == "" {
		return fmt.Errorf("TMDB API key is required. Set it via --api-key flag or TMDB_API_KEY environment variable")
	}

	// Initialize services
	tmdbService, err := services.NewTMDBService(apiKey)
	if err != nil {
		return fmt.Errorf("failed to initialize TMDB service: %w", err)
	}

	fileService := services.NewFileService()

	// Scan for video files
	fmt.Printf("Scanning directory: %s\n", applyInputDir)
	videoFiles, err := fileService.ScanDirectory(applyInputDir, applyRecursive)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(videoFiles) == 0 {
		color.Yellow("No video files found in the specified directory.")
		return nil
	}

	fmt.Printf("Found %d video file(s)\n\n", len(videoFiles))

	// Color setup
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	// Process each video file
	var results []models.RenameResult
	for i, videoFile := range videoFiles {
		fmt.Printf("Processing [%d/%d]: %s\n", i+1, len(videoFiles), cyan.Sprint(videoFile.OriginalName))

		result := processVideoFileForApply(videoFile, tmdbService, fileService, applyMediaType)
		results = append(results, result)

		if result.Success {
			fmt.Print("  ")
			green.Print("✓ ")
			fmt.Printf("Would rename to: %s\n", green.Sprint(result.NewFileName))
		} else {
			fmt.Print("  ")
			red.Print("✗ ")
			fmt.Printf("Error: %v\n", red.Sprint(result.Error))
		}
		fmt.Println()
	}

	// Summary
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	fmt.Println(color.HiBlackString("─────────────────────────────────────────────────────────────"))
	fmt.Printf("Summary: ")
	green.Printf("%d successful", successCount)
	fmt.Print(", ")
	if len(results)-successCount > 0 {
		red.Printf("%d failed", len(results)-successCount)
	} else {
		fmt.Print("0 failed")
	}
	fmt.Printf(", %d total\n", len(results))

	// Perform actual renames if not dry-run
	if !applyDryRun {
		if applyInteractive {
			fmt.Print("\n")
			yellow.Print("Proceed with renaming? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		fmt.Println("\nApplying renames...")
		for _, result := range results {
			if result.Success {
				oldPath := result.VideoFile.Path
				dir := filepath.Dir(oldPath)
				newPath := filepath.Join(dir, result.NewFileName)

				if err := fileService.RenameFile(oldPath, newPath); err != nil {
					fmt.Print("  ")
					red.Print("✗ ")
					fmt.Printf("Failed to rename %s: %v\n", result.VideoFile.OriginalName, err)
				} else {
					fmt.Print("  ")
					green.Print("✓ ")
					fmt.Printf("Renamed: %s\n", result.NewFileName)
				}
			}
		}
	}

	return nil
}

// processVideoFileForApply processes a single video file and returns a rename result
func processVideoFileForApply(videoFile models.VideoFile, tmdbService *services.TMDBService, fileService *services.FileService, mediaTypeOverride string) models.RenameResult {
	result := models.RenameResult{
		VideoFile: videoFile,
		Success:   false,
	}

	// Clean the filename for searching
	cleanName := services.CleanTitle(videoFile.OriginalName)
	year := services.ExtractYear(videoFile.OriginalName)

	// Determine media type
	detectedType := videoFile.MediaType
	if mediaTypeOverride != "auto" {
		switch mediaTypeOverride {
		case "movie":
			detectedType = models.MediaTypeMovie
		case "tv":
			detectedType = models.MediaTypeTVShow
		}
	}

	switch detectedType {
	case models.MediaTypeMovie:
		movie, err := tmdbService.SearchMovie(cleanName, year)
		if err != nil {
			result.Error = err
			return result
		}

		result.MediaInfo = movie
		result.NewFileName = fileService.GenerateMovieFileName(movie, videoFile.Path)
		result.Success = true

	case models.MediaTypeTVShow:
		// For TV shows, we need to extract season and episode numbers
		season, episode := extractSeasonEpisodeForApply(videoFile.OriginalName)
		if season == 0 || episode == 0 {
			result.Error = fmt.Errorf("could not extract season/episode information")
			return result
		}

		show, err := tmdbService.SearchTVShow(cleanName, year)
		if err != nil {
			result.Error = err
			return result
		}

		showId, err := strconv.Atoi(show.Database.IMDBId)
		if err != nil {
			result.Error = fmt.Errorf("invalid show ID: %w", err)
			return result
		}

		episodeInfo, err := tmdbService.GetEpisode(showId, season, episode)
		if err != nil {
			result.Error = err
			return result
		}

		result.MediaInfo = map[string]interface{}{
			"show":    show,
			"episode": episodeInfo,
		}
		result.NewFileName = fileService.GenerateTVShowFileName(show, episodeInfo, videoFile.Path)
		result.Success = true
	}

	return result
}

// extractSeasonEpisodeForApply extracts season and episode numbers from filename
func extractSeasonEpisodeForApply(filename string) (season, episode int) {
	filename = strings.ToLower(filename)

	// Simple pattern matching for S##E## format
	for i, char := range filename {
		if char == 's' && i < len(filename)-4 {
			// Look for S##E## pattern
			remaining := filename[i:]
			if len(remaining) >= 6 { // At least "s##e##"
				if remaining[1] >= '0' && remaining[1] <= '9' {
					seasonEnd := 2
					if remaining[2] >= '0' && remaining[2] <= '9' {
						seasonEnd = 3
					}

					if seasonStr := remaining[1:seasonEnd]; len(seasonStr) >= 1 {
						if s, err := strconv.Atoi(seasonStr); err == nil {
							season = s
						}
					}

					eIndex := strings.Index(remaining, "e")
					if eIndex != -1 && eIndex < len(remaining)-1 {
						episodeEnd := eIndex + 2
						if eIndex+2 < len(remaining) && remaining[eIndex+2] >= '0' && remaining[eIndex+2] <= '9' {
							episodeEnd = eIndex + 3
						}

						if episodeStr := remaining[eIndex+1 : episodeEnd]; len(episodeStr) >= 1 {
							if e, err := strconv.Atoi(episodeStr); err == nil {
								episode = e
								break
							}
						}
					}
				}
			}
		}
	}

	return season, episode
}
