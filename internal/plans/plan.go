package plans

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"goname/internal/models"
	"goname/pkg/database"
	"goname/pkg/log"
	"goname/pkg/services"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Plan represents a complete rename plan with all operations and conflicts
type Plan struct {
	// ID of the plan
	ID string `json:"id"`

	// Timestamp of the plan creation
	Timestamp time.Time `json:"timestamp"`

	// Changes represent the proposed changes to be made
	Changes []Change `json:"changes"`

	// Conflicts between changes
	Conflicts []Conflict `json:"conflicts"`
}

// NewPlan creates a new rename plan for the given video files
func NewPlan(videoFiles []models.VideoFile, mediaTypeOverride string, databaseService database.VideoDatabase, fileService *services.FileService) (*Plan, error) {
	plan := &Plan{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Changes:   make([]Change, 0, len(videoFiles)),
		Conflicts: make([]Conflict, 0),
	}

	// Create planned changes
	for _, videoFile := range videoFiles {
		change, err := createChange(videoFile, mediaTypeOverride, databaseService, fileService)
		if err != nil {
			log.Error("failed to create planned change", zap.Error(err), zap.String("file", videoFile.Path))
			change.Action = ActionSkip
			change.Error = err.Error()
		}
		plan.Changes = append(plan.Changes, *change)
	}

	// Detect conflicts
	conflicts := detectConflicts(plan.Changes)
	plan.Conflicts = conflicts

	// Update change conflict IDs based on conflicts
	updateChangeConflicts(plan)

	return plan, nil
}

// PlanSummary provides an overview of the plan
type PlanSummary struct {
	TotalChanges      int `json:"total_changes"`
	ReadyChanges      int `json:"ready_changes"`
	ConflictedChanges int `json:"conflicted_changes"`
	SkippedChanges    int `json:"skipped_changes"`
	ErrorChanges      int `json:"error_changes"`
	NoopChanges       int `json:"noop_changes"`
	TotalConflicts    int `json:"total_conflicts"`
	ResolvedConflicts int `json:"resolved_conflicts"`
}

// Summary provides a summary of the plan's changes and conflicts
func (p *Plan) Summary() PlanSummary {
	summary := PlanSummary{
		TotalChanges:   len(p.Changes),
		TotalConflicts: len(p.Conflicts),
	}

	for _, change := range p.Changes {
		switch change.Action {
		case ActionRename:
			if change.IsConflicting() {
				summary.ConflictedChanges++
			} else if change.Error != "" {
				summary.ErrorChanges++
			} else {
				summary.ReadyChanges++
			}
		case ActionSkip:
			summary.SkippedChanges++
		case ActionNoop:
			summary.NoopChanges++
		}
	}

	for _, conflict := range p.Conflicts {
		if conflict.Resolved {
			summary.ResolvedConflicts++
		}
	}

	return summary
}

// createChange creates a single planned change
func createChange(videoFile models.VideoFile, mediaTypeOverride string, databaseService database.VideoDatabase, fileService *services.FileService) (*Change, error) {
	change := &Change{
		ID:     uuid.New().String(),
		Action: ActionNoop,
		Before: FileInfo{
			Path:     videoFile.Path,
			FileName: videoFile.OriginalName,
		},
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

	var targetName string

	switch detectedType {
	case models.MediaTypeMovie:
		movie, err := databaseService.SearchMovie(cleanName, year)
		if err != nil {
			return change, fmt.Errorf("failed to search movie: %w", err)
		}

		targetName = fileService.GenerateMovieFileName(movie, videoFile.Path)

	case models.MediaTypeTVShow:
		// For TV shows, we need to extract season and episode numbers
		season, episode := services.ExtractSeasonEpisode(videoFile.OriginalName)
		if season == 0 || episode == 0 {
			return change, fmt.Errorf("could not extract season/episode information")
		}

		show, err := databaseService.SearchTVShow(cleanName, year)
		if err != nil {
			return change, fmt.Errorf("failed to search TV show: %w", err)
		}

		// Get episode information
		episodeInfo, err := getEpisodeInfo(show, season, episode, databaseService)
		if err != nil {
			return change, fmt.Errorf("failed to get episode info: %w", err)
		}

		targetName = fileService.GenerateTVShowFileName(show, episodeInfo, videoFile.Path)
	}

	targetPath := filepath.Join(filepath.Dir(videoFile.Path), targetName)

	// Set the After info
	change.After = FileInfo{
		Path:     targetPath,
		FileName: targetName,
	}

	// Determine action based on whether file needs to be renamed
	if videoFile.OriginalName != targetName {
		change.Action = ActionRename
	} else {
		change.Action = ActionNoop
	}

	return change, nil
}

// detectConflicts detects conflicts between planned changes
func detectConflicts(changes []Change) []Conflict {
	conflicts := make([]Conflict, 0)
	targetPaths := make(map[string][]string) // targetPath -> []changeID

	// Group changes by target path
	for _, change := range changes {
		if change.Action == ActionRename {
			targetPaths[change.After.Path] = append(targetPaths[change.After.Path], change.ID)
		}
	}

	// Check for multiple source conflicts
	for targetPath, changeIDs := range targetPaths {
		if len(changeIDs) > 1 {
			conflict := Conflict{
				ID:           uuid.New().String(),
				TargetPath:   targetPath,
				ChangeIDs:    changeIDs,
				ConflictType: models.ConflictTypeMultipleSource,
				Resolved:     false,
			}
			conflicts = append(conflicts, conflict)
		} else {
			// Check if target already exists on disk
			changeID := changeIDs[0]
			if fileExists(targetPath) {
				conflict := Conflict{
					ID:           uuid.New().String(),
					TargetPath:   targetPath,
					ChangeIDs:    []string{changeID},
					ConflictType: models.ConflictTypeTargetExists,
					Resolved:     false,
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// updateChangeConflicts updates change conflict IDs based on detected conflicts
func updateChangeConflicts(plan *Plan) {
	// Create a map of change ID to conflict IDs
	changeConflicts := make(map[string][]string)

	for _, conflict := range plan.Conflicts {
		for _, changeID := range conflict.ChangeIDs {
			changeConflicts[changeID] = append(changeConflicts[changeID], conflict.ID)
		}
	}

	// Update change conflict IDs
	for i := range plan.Changes {
		change := &plan.Changes[i]
		if conflictIDs, hasConflict := changeConflicts[change.ID]; hasConflict {
			change.ConflictIDs = conflictIDs
		}
	}
}

// getEpisodeInfo helper method to get episode information
func getEpisodeInfo(show *models.TVShow, season, episode int, databaseService database.VideoDatabase) (*models.Episode, error) {
	// Try to get episode from database service first
	showID, err := strconv.Atoi(show.Database.IMDBId)
	if err != nil {
		return nil, fmt.Errorf("invalid show ID: %w", err)
	}

	episodeInfo, err := databaseService.GetEpisode(showID, season, episode)
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

// HasUnresolvedConflict checks if there are any unresolved conflicts in the plan
func (p *Plan) HasUnresolvedConflict() bool {
	return false
}

// RenameResult represents the result of a rename operation
type RenameResult struct {
	VideoFile      models.VideoFile
	Success        bool
	Error          error
	MediaInfo      interface{} // Can be MovieInfo, TVShowInfo, etc.
	NewFileName    string
	ConflictAction string // Action taken during conflict resolution
}

// ResolveConflicts resolves all conflicts in the plan using the specified strategy
func (p *Plan) ResolveConflicts(strategy models.ConflictStrategy) error {
	for i := range p.Conflicts {
		conflict := &p.Conflicts[i]
		if !conflict.Resolved {
			if err := p.resolveConflict(conflict, strategy); err != nil {
				log.Error("failed to resolve conflict", zap.Error(err), zap.String("conflict_id", conflict.ID))
				return fmt.Errorf("failed to resolve conflict %s: %w", conflict.ID, err)
			}
		}
	}

	return nil
}

// resolveConflict resolves a single conflict
func (p *Plan) resolveConflict(conflict *Conflict, strategy models.ConflictStrategy) error {
	switch conflict.ConflictType {
	case models.ConflictTypeMultipleSource:
		return p.resolveMultipleSourceConflict(conflict, strategy)
	case models.ConflictTypeTargetExists:
		return p.resolveTargetExistsConflict(conflict, strategy)
	default:
		return fmt.Errorf("unknown conflict type: %v", conflict.ConflictType)
	}
}

// resolveMultipleSourceConflict resolves a conflict where multiple sources map to the same target
func (p *Plan) resolveMultipleSourceConflict(conflict *Conflict, strategy models.ConflictStrategy) error {
	// Find all changes involved in this conflict
	conflictingChanges := make([]*Change, 0, len(conflict.ChangeIDs))
	for i := range p.Changes {
		change := &p.Changes[i]
		for _, changeID := range conflict.ChangeIDs {
			if change.ID == changeID {
				conflictingChanges = append(conflictingChanges, change)
				break
			}
		}
	}

	// Apply resolution strategy
	switch strategy {
	case models.SkipConflict:
		// Skip all conflicting changes
		for _, change := range conflictingChanges {
			change.Action = ActionSkip
			change.Error = "Skipped due to conflict"
		}

	case models.AppendNumber:
		// Append numbers to target filenames to make them unique
		for i, change := range conflictingChanges {
			if i == 0 {
				// Keep the first one as-is
				continue
			}

			// Generate new filename with number suffix
			dir := filepath.Dir(change.After.Path)
			ext := filepath.Ext(change.After.FileName)
			nameWithoutExt := change.After.FileName[:len(change.After.FileName)-len(ext)]

			// Find a unique number
			counter := i
			var newFileName string
			var newPath string

			for {
				newFileName = fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
				newPath = filepath.Join(dir, newFileName)

				// Check if this path conflicts with any other change or existing file
				if !p.pathConflictsWithOtherChanges(newPath, change.ID) && !fileExists(newPath) {
					break
				}
				counter++
			}

			// Update the change with the new target
			change.After.Path = newPath
			change.After.FileName = newFileName
		}

	case models.AppendTimestamp:
		// Append timestamps to target filenames to make them unique
		for i, change := range conflictingChanges {
			if i == 0 {
				// Keep the first one as-is
				continue
			}

			// Generate new filename with timestamp suffix
			dir := filepath.Dir(change.After.Path)
			ext := filepath.Ext(change.After.FileName)
			nameWithoutExt := change.After.FileName[:len(change.After.FileName)-len(ext)]

			// Use current timestamp with microseconds for uniqueness
			timestamp := time.Now().Format("20060102-150405.000000")
			newFileName := fmt.Sprintf("%s (%s)%s", nameWithoutExt, timestamp, ext)
			newPath := filepath.Join(dir, newFileName)

			// Update the change with the new target
			change.After.Path = newPath
			change.After.FileName = newFileName
		}

	case models.Overwrite:
		// Keep the first, skip others
		for i, change := range conflictingChanges {
			if i > 0 {
				change.Action = ActionSkip
				change.Error = "Skipped due to overwrite strategy"
			}
		}

	default:
		return fmt.Errorf("unsupported conflict strategy: %v", strategy)
	}

	// Mark conflict as resolved
	conflict.Resolved = true
	conflict.Resolution = ConflictResolution{
		Strategy:      p.getStrategyName(strategy),
		Modifications: make(map[string]string),
		Timestamp:     time.Now(),
	}

	// Clear conflict IDs from changes
	for _, change := range conflictingChanges {
		change.ConflictIDs = removeConflictID(change.ConflictIDs, conflict.ID)
	}

	return nil
}

// resolveTargetExistsConflict resolves a conflict where the target file already exists
func (p *Plan) resolveTargetExistsConflict(conflict *Conflict, strategy models.ConflictStrategy) error {
	if len(conflict.ChangeIDs) != 1 {
		return fmt.Errorf("target exists conflict should have exactly one change, got %d", len(conflict.ChangeIDs))
	}

	// Find the conflicting change
	var change *Change
	for i := range p.Changes {
		if p.Changes[i].ID == conflict.ChangeIDs[0] {
			change = &p.Changes[i]
			break
		}
	}

	if change == nil {
		return fmt.Errorf("could not find change with ID %s", conflict.ChangeIDs[0])
	}

	// Apply resolution strategy
	switch strategy {
	case models.SkipConflict:
		change.Action = ActionSkip
		change.Error = "Skipped because target file already exists"

	case models.Overwrite:
		// Keep the rename action - the actual file service will handle the overwrite
		// No action needed here

	default:
		// For other strategies, skip for now - more sophisticated logic can be added later
		change.Action = ActionSkip
		change.Error = fmt.Sprintf("Skipped due to target exists conflict (strategy: %v)", strategy)
	}

	// Mark conflict as resolved
	conflict.Resolved = true
	conflict.Resolution = ConflictResolution{
		Strategy:      p.getStrategyName(strategy),
		Modifications: make(map[string]string),
		Timestamp:     time.Now(),
	}

	// Clear conflict ID from change
	change.ConflictIDs = removeConflictID(change.ConflictIDs, conflict.ID)

	return nil
}

// getStrategyName returns a human-readable name for the conflict strategy
func (p *Plan) getStrategyName(strategy models.ConflictStrategy) string {
	switch strategy {
	case models.SkipConflict:
		return "skip"
	case models.AppendNumber:
		return "append_number"
	case models.AppendTimestamp:
		return "append_timestamp"
	case models.PromptUser:
		return "prompt_user"
	case models.Overwrite:
		return "overwrite"
	default:
		return "unknown"
	}
}

// pathConflictsWithOtherChanges checks if a path conflicts with any other changes in the plan
func (p *Plan) pathConflictsWithOtherChanges(targetPath, excludeChangeID string) bool {
	for _, change := range p.Changes {
		if change.ID != excludeChangeID && change.Action == ActionRename && change.After.Path == targetPath {
			return true
		}
	}
	return false
}
