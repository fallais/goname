package plan

import (
	"fmt"

	"goname/internal/cmd/common"
	"goname/internal/plans"
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
	planService := plans.NewPlanService(databaseService, fileService)

	// Create plan conflict resolver
	planConflictResolver := plans.NewPlanConflictResolver(conflictStrategy)

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
}
