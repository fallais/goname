package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"goname/internal/models"
	"goname/pkg/services"
)

var (
	inputDir    string
	recursive   bool
	dryRun      bool
	tmdbAPIKey  string
	mediaType   string
	interactive bool
)

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename video files using TMDB data",
	Long: `Rename video files in a directory by fetching information from The Movie Database (TMDB).

The command will scan for video files and attempt to match them with movies or TV shows
from TMDB, then rename them using a standardized format.

Examples:
  # Rename movies in current directory
  goname rename --type movie --api-key YOUR_API_KEY
  
  # Rename TV shows recursively with dry-run
  goname rename --dir /path/to/shows --type tv --recursive --dry-run --api-key YOUR_API_KEY
  
  # Interactive mode for manual confirmation
  goname rename --interactive --api-key YOUR_API_KEY`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := runRename(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	renameCmd.Flags().StringVarP(&inputDir, "dir", "d", ".", "Directory to scan for video files")
	renameCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Scan directories recursively")
	renameCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be renamed without actually renaming")
	renameCmd.Flags().StringVar(&tmdbAPIKey, "api-key", "", "TMDB API key (can also be set via TMDB_API_KEY env var)")
	renameCmd.Flags().StringVarP(&mediaType, "type", "t", "auto", "Media type: movie, tv, or auto")
	renameCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode for manual confirmation")

	// Bind flags to viper
	viper.BindPFlag("tmdb.api_key", renameCmd.Flags().Lookup("api-key"))
	viper.BindEnv("tmdb.api_key", "TMDB_API_KEY")
}

func runRename() error {
	// Get API key from flag, config, or environment
	apiKey := tmdbAPIKey
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
	fmt.Printf("Scanning directory: %s\n", inputDir)
	videoFiles, err := fileService.ScanDirectory(inputDir, recursive)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(videoFiles) == 0 {
		fmt.Println("No video files found in the specified directory.")
		return nil
	}

	fmt.Printf("Found %d video file(s)\n\n", len(videoFiles))

	// Process each video file
	var results []models.RenameResult
	for i, videoFile := range videoFiles {
		fmt.Printf("Processing [%d/%d]: %s\n", i+1, len(videoFiles), videoFile.OriginalName)

		result := processVideoFile(videoFile, tmdbService, fileService)
		results = append(results, result)

		if result.Success {
			fmt.Printf("✓ Would rename to: %s\n", result.NewFileName)
		} else {
			fmt.Printf("✗ Error: %v\n", result.Error)
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

	fmt.Printf("Summary: %d successful, %d failed\n", successCount, len(results)-successCount)

	// Perform actual renames if not dry-run
	if !dryRun {
		if interactive {
			fmt.Print("Proceed with renaming? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		fmt.Println("\nRenaming files...")
		for _, result := range results {
			if result.Success {
				oldPath := result.VideoFile.Path
				dir := filepath.Dir(oldPath)
				newPath := filepath.Join(dir, result.NewFileName)

				if err := fileService.RenameFile(oldPath, newPath); err != nil {
					fmt.Printf("✗ Failed to rename %s: %v\n", result.VideoFile.OriginalName, err)
				} else {
					fmt.Printf("✓ Renamed: %s\n", result.NewFileName)
				}
			}
		}
	}

	return nil
}

func processVideoFile(videoFile models.VideoFile, tmdbService *services.TMDBService, fileService *services.FileService) models.RenameResult {
	result := models.RenameResult{
		VideoFile: videoFile,
		Success:   false,
	}

	// Clean the filename for searching
	cleanName := services.CleanTitle(videoFile.OriginalName)
	year := services.ExtractYear(videoFile.OriginalName)

	// Determine media type
	detectedType := videoFile.MediaType
	if mediaType != "auto" {
		switch mediaType {
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
		season, episode := extractSeasonEpisode(videoFile.OriginalName)
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

// extractSeasonEpisode extracts season and episode numbers from filename
func extractSeasonEpisode(filename string) (season, episode int) {
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
