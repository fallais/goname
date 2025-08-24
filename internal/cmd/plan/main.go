package plan

import (
	"fmt"
	"path/filepath"
	"strings"

	"goname/internal/cmd/common"
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
	default:
		log.Fatal("unsupported database type", zap.String("db", viper.GetString("db")))
	}

	// Create conflict resolver
	conflictStrategy, err := services.ParseConflictStrategy(viper.GetString("conflict"))
	if err != nil {
		log.Fatal("failed to parse conflict strategy", zap.Error(err))
	}
	conflictResolver := services.NewConflictResolver(conflictStrategy)

	// Create file service
	fileService := services.NewFileService("", "", conflictResolver)

	// Create plan service
	planService := services.NewPlanService(databaseService, fileService)

	// Create plan conflict resolver
	planConflictResolver := services.NewPlanConflictResolver(conflictStrategy)

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

	// Create the plan
	plan, err := planService.CreatePlan(videoFiles, viper.GetString("type"))
	if err != nil {
		log.Fatal("failed to create plan", zap.Error(err))
	}

	// Resolve conflicts
	if len(plan.Conflicts) > 0 {
		log.Debug("conflicts detected", zap.Int("nb_conflicts", len(plan.Conflicts)))
		if err := planConflictResolver.ResolvePlanConflicts(plan); err != nil {
			log.Fatal("failed to resolve conflicts", zap.Error(err))
		}
	}

	// Display results
	displayPlanResults(plan)
}

// displayPlanResults displays the results of a rename plan
func displayPlanResults(plan *models.Plan) {
	// Color setup
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow)
	blue := color.New(color.FgBlue)

	alreadyCorrectCount := 0
	needsRenameCount := 0
	errorCount := 0
	skippedCount := 0

	fmt.Println("GoName will perform the following actions:")
	fmt.Println()

	for _, operation := range plan.Operations {
		switch operation.Status {
		case models.OperationStatusReady:
			// Check if the current filename already matches the proposed new filename
			currentBaseName := strings.TrimSuffix(operation.VideoFile.OriginalName, filepath.Ext(operation.VideoFile.OriginalName))
			proposedBaseName := strings.TrimSuffix(operation.TargetName, filepath.Ext(operation.TargetName))

			if currentBaseName == proposedBaseName {
				// File is already correctly named
				alreadyCorrectCount++
				green.Printf("%s\n", operation.VideoFile.OriginalName)
			} else {
				// File needs to be renamed
				needsRenameCount++
				fmt.Printf("%s → %s\n", operation.VideoFile.OriginalName, yellow.Sprint(operation.TargetName))
			}

		case models.OperationStatusSkipped:
			skippedCount++
			fmt.Printf("%s: %s\n", operation.VideoFile.OriginalName, blue.Sprint("SKIPPED"))
			if operation.Error != "" {
				fmt.Printf("    Reason: %s\n", operation.Error)
			}

		case models.OperationStatusError:
			errorCount++
			fmt.Printf("%s: %v\n", operation.VideoFile.OriginalName, red.Sprint(operation.Error))

		case models.OperationStatusConflicted:
			// This should not happen after conflict resolution
			errorCount++
			fmt.Printf("%s: %s\n", operation.VideoFile.OriginalName, red.Sprint("UNRESOLVED CONFLICT"))
		}

		if AddCarriageReturn {
			fmt.Println()
		}
	}

	// Summary
	common.DisplayPlanResults(plan)
}
