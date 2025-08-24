package plan

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"goname/internal/models"
	"goname/pkg/database"
	"goname/pkg/database/tmdb"
	"goname/pkg/log"
	"goname/pkg/services"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const AddCarriageReturn = false

func Run(cmd *cobra.Command, args []string) {
	// Initialize services
	var databaseService database.VideoDatabase
	switch viper.GetString("db") {
	case "tmdb":
		tmdbService, err := tmdb.New(viper.GetString("tmdb.api_key"))
		if err != nil {
			log.Fatal("failed to initialize TMDB service", zap.Error(err))
		}

		databaseService = tmdbService
	case "tvdb":
		/* tvdbService, err := tvdb.New(viper.GetString("tvdb.api_key"))
		if err != nil {
			log.Fatalf("failed to initialize TVDB service: %v", err)
		} */
	default:
		log.Fatal("unsupported database type", zap.String("db", viper.GetString("db")))
	}

	fileService := services.NewFileService()

	// Scan for video files
	fmt.Printf("Scanning directory: %s\n", viper.GetString("dir"))
	videoFiles, err := fileService.ScanDirectory(viper.GetString("dir"), viper.GetBool("recursive"))
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
	yellow := color.New(color.FgYellow)

	// Process each video file
	var results []models.RenameResult
	alreadyCorrectCount := 0
	needsRenameCount := 0
	errorCount := 0

	for _, videoFile := range videoFiles {
		result := processVideoFileForPlan(videoFile, databaseService, fileService, viper.GetString("type"))
		results = append(results, result)

		if result.Success {
			// Check if the current filename already matches the proposed new filename
			currentBaseName := strings.TrimSuffix(videoFile.OriginalName, filepath.Ext(videoFile.OriginalName))
			proposedBaseName := strings.TrimSuffix(result.NewFileName, filepath.Ext(result.NewFileName))

			if currentBaseName == proposedBaseName {
				// File is already correctly named
				alreadyCorrectCount++
				fmt.Print("  ")
				green.Printf("%s\n", videoFile.OriginalName)
			} else {
				// File needs to be renamed
				needsRenameCount++
				fmt.Print("  ")
				fmt.Printf("%s → %s\n", videoFile.OriginalName, yellow.Sprint(result.NewFileName))
			}
		} else {
			errorCount++
			fmt.Print("  ")
			fmt.Printf("%s: %v\n", videoFile.OriginalName, red.Sprint(result.Error))
		}

		if AddCarriageReturn {
			fmt.Println()
		}
	}

	// Summary
	fmt.Println()
	fmt.Println(color.HiBlackString("─────────────────────────────────────────────────────────────"))
	fmt.Printf("Plan Summary: ")
	green.Printf("%d already correct", alreadyCorrectCount)
	fmt.Print(", ")
	yellow.Printf("%d to rename", needsRenameCount)
	fmt.Print(", ")
	if errorCount > 0 {
		red.Printf("%d errors", errorCount)
	} else {
		fmt.Print("0 errors")
	}
	fmt.Printf(", %d total\n", len(results))

	if needsRenameCount > 0 {
		fmt.Println()
		yellow.Println("To apply these changes, run: goname apply")
	}
}

// processVideoFileForPlan processes a single video file and returns a rename result
func processVideoFileForPlan(videoFile models.VideoFile, tmdbService database.VideoDatabase, fileService *services.FileService, mediaTypeOverride string) models.RenameResult {
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

		episodeID, err := strconv.Atoi(show.Database.IMDBId)
		if err != nil {
			result.Error = fmt.Errorf("invalid episode ID: %w", err)
			return result
		}

		episodeInfo, err := tmdbService.GetEpisode(episodeID, season, episode)
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
