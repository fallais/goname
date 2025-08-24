package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"goname/internal/models"
	"goname/pkg/database"
	"goname/pkg/log"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// PlanService handles the creation and management of rename plans
type PlanService struct {
	databaseService database.VideoDatabase
	fileService     *FileService
}

// NewPlanService creates a new PlanService
func NewPlanService(databaseService database.VideoDatabase, fileService *FileService) *PlanService {
	return &PlanService{
		databaseService: databaseService,
		fileService:     fileService,
	}
}

// CreatePlan creates a new rename plan for the given video files
func (ps *PlanService) CreatePlan(videoFiles []models.VideoFile, mediaTypeOverride string) (*models.Plan, error) {
	plan := &models.Plan{
		ID:         uuid.New().String(),
		Timestamp:  time.Now(),
		Operations: make([]models.PlannedOperation, 0, len(videoFiles)),
		Conflicts:  make([]models.Conflict, 0),
		Resolved:   false,
	}

	// First pass: create planned operations
	for _, videoFile := range videoFiles {
		operation, err := ps.createPlannedOperation(videoFile, mediaTypeOverride)
		if err != nil {
			log.Error("failed to create planned operation", zap.Error(err), zap.String("file", videoFile.Path))
			operation.Status = models.OperationStatusError
			operation.Error = err.Error()
		}
		plan.Operations = append(plan.Operations, *operation)
	}

	// Second pass: detect conflicts
	conflicts := ps.detectConflicts(plan.Operations)
	plan.Conflicts = conflicts

	// Update operation status based on conflicts
	ps.updateOperationStatuses(plan)

	// Calculate summary
	plan.Summary = ps.calculateSummary(plan)

	return plan, nil
}

// createPlannedOperation creates a single planned operation
func (ps *PlanService) createPlannedOperation(videoFile models.VideoFile, mediaTypeOverride string) (*models.PlannedOperation, error) {
	operation := &models.PlannedOperation{
		ID:        uuid.New().String(),
		VideoFile: videoFile,
		Status:    models.OperationStatusPending,
	}

	// Clean the filename for searching
	cleanName := CleanTitle(videoFile.OriginalName)
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
		movie, err := ps.databaseService.SearchMovie(cleanName, year)
		if err != nil {
			return operation, fmt.Errorf("failed to search movie: %w", err)
		}

		operation.MediaInfo = movie
		operation.TargetName = ps.fileService.GenerateMovieFileName(movie, videoFile.Path)
		operation.TargetPath = filepath.Join(filepath.Dir(videoFile.Path), operation.TargetName)
		operation.Status = models.OperationStatusReady

	case models.MediaTypeTVShow:
		// For TV shows, we need to extract season and episode numbers
		season, episode := extractSeasonEpisode(videoFile.OriginalName)
		if season == 0 || episode == 0 {
			return operation, fmt.Errorf("could not extract season/episode information")
		}

		show, err := ps.databaseService.SearchTVShow(cleanName, year)
		if err != nil {
			return operation, fmt.Errorf("failed to search TV show: %w", err)
		}

		// Get episode information (assuming you have this method)
		// You might need to adjust this based on your actual database interface
		episodeInfo, err := ps.getEpisodeInfo(show, season, episode)
		if err != nil {
			return operation, fmt.Errorf("failed to get episode info: %w", err)
		}

		operation.MediaInfo = map[string]interface{}{
			"show":    show,
			"episode": episodeInfo,
		}
		operation.TargetName = ps.fileService.GenerateTVShowFileName(show, episodeInfo, videoFile.Path)
		operation.TargetPath = filepath.Join(filepath.Dir(videoFile.Path), operation.TargetName)
		operation.Status = models.OperationStatusReady
	}

	return operation, nil
}

// detectConflicts detects conflicts between planned operations
func (ps *PlanService) detectConflicts(operations []models.PlannedOperation) []models.Conflict {
	conflicts := make([]models.Conflict, 0)
	targetPaths := make(map[string][]string) // targetPath -> []operationID

	// Group operations by target path
	for _, op := range operations {
		if op.Status == models.OperationStatusReady {
			targetPaths[op.TargetPath] = append(targetPaths[op.TargetPath], op.ID)
		}
	}

	// Check for multiple source conflicts
	for targetPath, operationIDs := range targetPaths {
		if len(operationIDs) > 1 {
			conflict := models.Conflict{
				ID:           uuid.New().String(),
				TargetPath:   targetPath,
				OperationIDs: operationIDs,
				ConflictType: models.ConflictTypeMultipleSource,
				Resolved:     false,
			}
			conflicts = append(conflicts, conflict)
		} else {
			// Check if target already exists on disk
			operationID := operationIDs[0]
			if fileExists(targetPath) {
				conflict := models.Conflict{
					ID:           uuid.New().String(),
					TargetPath:   targetPath,
					OperationIDs: []string{operationID},
					ConflictType: models.ConflictTypeTargetExists,
					Resolved:     false,
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// updateOperationStatuses updates operation statuses based on detected conflicts
func (ps *PlanService) updateOperationStatuses(plan *models.Plan) {
	// Create a map of operation ID to conflict IDs
	operationConflicts := make(map[string][]string)

	for _, conflict := range plan.Conflicts {
		for _, opID := range conflict.OperationIDs {
			operationConflicts[opID] = append(operationConflicts[opID], conflict.ID)
		}
	}

	// Update operation statuses
	for i := range plan.Operations {
		op := &plan.Operations[i]
		if conflictIDs, hasConflict := operationConflicts[op.ID]; hasConflict {
			op.Status = models.OperationStatusConflicted
			op.ConflictIDs = conflictIDs
		}
	}
}

// calculateSummary calculates the plan summary
func (ps *PlanService) calculateSummary(plan *models.Plan) models.PlanSummary {
	summary := models.PlanSummary{
		TotalOperations: len(plan.Operations),
		TotalConflicts:  len(plan.Conflicts),
	}

	for _, op := range plan.Operations {
		switch op.Status {
		case models.OperationStatusReady:
			summary.ReadyOperations++
		case models.OperationStatusConflicted:
			summary.ConflictedOperations++
		case models.OperationStatusSkipped:
			summary.SkippedOperations++
		case models.OperationStatusError:
			summary.ErrorOperations++
		}
	}

	for _, conflict := range plan.Conflicts {
		if conflict.Resolved {
			summary.ResolvedConflicts++
		}
	}

	return summary
}

// getEpisodeInfo helper method to get episode information
func (ps *PlanService) getEpisodeInfo(show *models.TVShow, season, episode int) (*models.Episode, error) {
	// Try to get episode from database service first
	showID, err := strconv.Atoi(show.Database.IMDBId)
	if err != nil {
		return nil, fmt.Errorf("invalid show ID: %w", err)
	}

	episodeInfo, err := ps.databaseService.GetEpisode(showID, season, episode)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode from database: %w", err)
	}

	return episodeInfo, nil
}

// fileExists checks if a file exists at the given path
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
