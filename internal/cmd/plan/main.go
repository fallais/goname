package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	// Create file service with conflict resolution
	conflictStrategy := services.ParseConflictStrategy(viper.GetString("conflict"))

	conflictResolver := services.NewConflictResolver(conflictStrategy)
	fileService := services.NewFileServiceWithConflictResolver(conflictResolver)

	// Scan for video files
	fmt.Printf("Scanning directory: %s\n", viper.GetString("dir"))
	fmt.Printf("Conflict resolution strategy: %s\n", conflictResolver.GetStrategyName())
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

	// First pass: generate all proposed filenames
	for _, videoFile := range videoFiles {
		result := processVideoFileForPlan(videoFile, databaseService, fileService, viper.GetString("type"))
		results = append(results, result)
	}

	// Second pass: detect and resolve conflicts
	targetPaths := make(map[string][]int) // map[targetPath][]resultIndex
	for i, result := range results {
		if result.Success && result.NewFileName != "" {
			// Build the full target path
			targetPath := filepath.Join(filepath.Dir(result.VideoFile.Path), result.NewFileName)
			targetPaths[targetPath] = append(targetPaths[targetPath], i)
		}
	}

	// Keep track of paths that are "taken" by files being renamed
	takenPaths := make(map[string]bool)

	// Apply conflict resolution to conflicting files
	for targetPath, indices := range targetPaths {
		if len(indices) > 1 {
			// Multiple files want the same target path - resolve conflicts
			for j, idx := range indices {
				if j == 0 {
					// First file gets the original name (unless it already exists on disk)
					if conflictResolver.CheckConflict(targetPath) {
						// File already exists on disk, resolve conflict
						conflictResult, err := conflictResolver.ResolveConflict(targetPath)
						if err == nil && !conflictResult.Skipped {
							results[idx].NewFileName = filepath.Base(conflictResult.ResolvedPath)
							results[idx].ConflictAction = conflictResult.Action
							takenPaths[conflictResult.ResolvedPath] = true
						}
					} else {
						// First file takes the original target path
						takenPaths[targetPath] = true
					}
				} else {
					// Subsequent files always need conflict resolution since first file took the path
					// We simulate a conflict by generating the resolved name directly
					conflictResult, err := simulateConflictResolution(conflictResolver, targetPath, true)
					if err == nil && !conflictResult.Skipped {
						results[idx].NewFileName = filepath.Base(conflictResult.ResolvedPath)
						results[idx].ConflictAction = conflictResult.Action
						takenPaths[conflictResult.ResolvedPath] = true
					} else if conflictResult != nil && conflictResult.Skipped {
						results[idx].Success = false
						results[idx].Error = fmt.Errorf("skipped due to conflict")
					}
				}
			}
		} else if len(indices) == 1 {
			// Single file, but check if target already exists on disk
			idx := indices[0]
			if conflictResolver.CheckConflict(targetPath) {
				conflictResult, err := conflictResolver.ResolveConflict(targetPath)
				if err == nil && !conflictResult.Skipped {
					results[idx].NewFileName = filepath.Base(conflictResult.ResolvedPath)
					results[idx].ConflictAction = conflictResult.Action
					takenPaths[conflictResult.ResolvedPath] = true
				} else if conflictResult != nil && conflictResult.Skipped {
					results[idx].Success = false
					results[idx].Error = fmt.Errorf("skipped due to conflict")
				}
			} else {
				takenPaths[targetPath] = true
			}
		}
	}

	// Display results
	for _, result := range results {

		if result.Success {
			// Check if the current filename already matches the proposed new filename
			currentBaseName := strings.TrimSuffix(result.VideoFile.OriginalName, filepath.Ext(result.VideoFile.OriginalName))
			proposedBaseName := strings.TrimSuffix(result.NewFileName, filepath.Ext(result.NewFileName))

			if currentBaseName == proposedBaseName {
				// File is already correctly named
				alreadyCorrectCount++
				fmt.Print("  ")
				green.Printf("%s\n", result.VideoFile.OriginalName)
			} else {
				// File needs to be renamed
				needsRenameCount++
				fmt.Print("  ")
				displayName := result.NewFileName
				if result.ConflictAction != "" {
					displayName += fmt.Sprintf(" (%s)", result.ConflictAction)
				}
				fmt.Printf("%s → %s\n", result.VideoFile.OriginalName, yellow.Sprint(displayName))
			}
		} else {
			errorCount++
			fmt.Print("  ")
			fmt.Printf("%s: %v\n", result.VideoFile.OriginalName, red.Sprint(result.Error))
		}

		if AddCarriageReturn {
			fmt.Println()
		}
	}

	// Summary
	printSummary(alreadyCorrectCount, needsRenameCount, errorCount, results)
}

func printSummary(alreadyCorrectCount, needsRenameCount, errorCount int, results []models.RenameResult) {
	// Color setup
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	fmt.Println(color.HiBlackString("─────────────────────────────────────────────────────────────"))
	fmt.Printf("Plan Summary: ")
	green.Printf("%d correct", alreadyCorrectCount)
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
	year := database.ExtractYear(videoFile.OriginalName)

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

// simulateConflictResolution simulates what would happen during conflict resolution
func simulateConflictResolution(conflictResolver *services.ConflictResolver, targetPath string, forceConflict bool) (*services.ConflictResult, error) {
	if !forceConflict && !conflictResolver.CheckConflict(targetPath) {
		// No actual conflict exists, return as-is
		return &services.ConflictResult{
			ResolvedPath: targetPath,
			Action:       "none",
			Skipped:      false,
		}, nil
	}

	// Force conflict resolution by directly calling the appropriate strategy
	switch conflictResolver.GetStrategyName() {
	case "append_number":
		resolvedPath, err := appendNumber(targetPath)
		if err != nil {
			return nil, err
		}
		return &services.ConflictResult{
			ResolvedPath: resolvedPath,
			Action:       "append_number",
			Skipped:      false,
		}, nil
	case "append_timestamp":
		resolvedPath, err := appendTimestamp(targetPath)
		if err != nil {
			return nil, err
		}
		return &services.ConflictResult{
			ResolvedPath: resolvedPath,
			Action:       "append_timestamp",
			Skipped:      false,
		}, nil
	case "skip":
		return &services.ConflictResult{
			ResolvedPath: "",
			Action:       "skipped",
			Skipped:      true,
		}, nil
	case "overwrite":
		return &services.ConflictResult{
			ResolvedPath: targetPath,
			Action:       "overwrite",
			Skipped:      false,
		}, nil
	default:
		// Fall back to the normal resolver
		return conflictResolver.ResolveConflict(targetPath)
	}
}

// appendNumber duplicates the logic from ConflictResolver
func appendNumber(targetPath string) (string, error) {
	ext := filepath.Ext(targetPath)
	base := strings.TrimSuffix(targetPath, ext)

	for i := 1; i <= 999; i++ {
		newPath := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath, nil
		}
	}
	return "", fmt.Errorf("unable to resolve conflict for %s after 999 attempts", targetPath)
}

// appendTimestamp duplicates the logic from ConflictResolver
func appendTimestamp(targetPath string) (string, error) {
	ext := filepath.Ext(targetPath)
	base := strings.TrimSuffix(targetPath, ext)
	timestamp := time.Now().Format("20060102_150405")
	newPath := fmt.Sprintf("%s_%s%s", base, timestamp, ext)
	return newPath, nil
}
