package apply

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

func Run(cmd *cobra.Command, args []string) {
	log.Debug("goname is starting", zap.String("command", "apply"))

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
	common.DisplayPlanResults(plan)

	if !viper.GetBool("auto-approve") {
		fmt.Println()
		fmt.Println()
		fmt.Println("Do you want to perform these actions?")
		fmt.Println("GoName will perform the actions described above.")
		fmt.Println("Only 'yes' will be accepted to approve.")
		fmt.Println()

		fmt.Print("Enter a value: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	// Initialize state service for tracking renames
	stateService, err := services.NewStateService()
	if err != nil {
		log.Error("failed to initialize state service", zap.Error(err))
		return
	}

	fmt.Println("\nApplying renames...")
	for _, result := range plan.Operations {
		if result.Status == models.OperationStatusReady {
			oldPath := result.VideoFile.Path
			dir := filepath.Dir(oldPath)
			newPath := filepath.Join(dir, result.TargetName)

			conflictResult, err := fileService.RenameFile(oldPath, newPath)
			if err != nil {
				fmt.Print("  ")
				common.Red.Print("✗ ")
				fmt.Printf("Failed to rename %s: %v\n", result.VideoFile.OriginalName, err)
			} else if conflictResult.Skipped {
				fmt.Print("  ")
				common.Yellow.Print("⚠ ")
				fmt.Printf("Skipped (conflict): %s\n", result.VideoFile.OriginalName)
			} else {
				fmt.Print("  ")
				common.Green.Print("✓ ")
				finalName := filepath.Base(conflictResult.ResolvedPath)
				fmt.Printf("Renamed: %s", finalName)
				if conflictResult.Action != "none" {
					fmt.Printf(" (%s)", conflictResult.Action)
				}
				fmt.Println()

				// Add to state

				if err := stateService.AddRenameOperation(
					oldPath,
					newPath,
					result.VideoFile.OriginalName,
					result.TargetName,
					result.MediaInfo,
				); err != nil {
					log.Error("failed to add rename to state", zap.Error(err))
					common.Yellow.Printf("    Warning: Failed to track rename in state\n")
				}

			}
		}
	}

}
