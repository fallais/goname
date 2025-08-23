package apply

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"goname/internal/models"
	"goname/pkg/log"
	"goname/pkg/services"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func Run(cmd *cobra.Command, args []string) {
	log.Debug("goname is starting", zap.String("command", "apply"))

	// Initialize services
	tmdbService, err := services.NewTMDBService(viper.GetString("tmdb.api_key"))
	if err != nil {
		log.Fatal("failed to initialize TMDB service", zap.Error(err))
	}

	fileService := services.NewFileService()

	// Scan for video files
	fmt.Printf("Scanning directory: %s\n", viper.GetString("apply.input_dir"))
	videoFiles, err := fileService.ScanDirectory(viper.GetString("apply.input_dir"), viper.GetBool("apply.recursive"))
	if err != nil {
		log.Fatal("failed to scan directory", zap.Error(err))
	}

	if len(videoFiles) == 0 {
		color.Yellow("No video files found in the specified directory.")
		return
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

		result := processVideoFileForApply(videoFile, tmdbService, fileService, viper.GetString("media.type"))
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
	if !viper.GetBool("apply.dry_run") && successCount > 0 {
		if viper.GetBool("apply.interactive") {
			fmt.Print("\n")
			yellow.Print("Proceed with renaming? (y/N): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				fmt.Println("Operation cancelled.")
				return
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
