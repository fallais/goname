package plans

import (
	"goname/internal/models"
	"time"
)

// Plan represents a complete rename plan with all operations and conflicts
type Plan struct {
	// ID of the plan
	ID string `json:"id"`

	// Timestamp of the plan creation
	Timestamp time.Time `json:"timestamp"`

	// Changes represent the proposed changes to be made
	Changes []Change `json:"changes"`

	// To be removed
	Operations []PlannedOperation `json:"operations"`

	Conflicts []Conflict `json:"conflicts"`

	// TO be removed, we should use HasUnresolvedConflict()
	Resolved bool `json:"resolved"`

	// TO be removed, we should use HasUnresolvedConflict()
	HasErrors bool `json:"has_errors"`
}

func (p *Plan) Summary() PlanSummary {
	summary := PlanSummary{
		TotalOperations: len(p.Operations),
		TotalConflicts:  len(p.Conflicts),
	}

	for _, op := range p.Operations {
		switch op.Status {
		case OperationStatusReady:
			summary.ReadyOperations++
		case OperationStatusConflicted:
			summary.ConflictedOperations++
		case OperationStatusSkipped:
			summary.SkippedOperations++
		case OperationStatusError:
			summary.ErrorOperations++
		}
	}

	for _, conflict := range p.Conflicts {
		if conflict.Resolved {
			summary.ResolvedConflicts++
		}
	}

	return summary
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

// PlannedOperation represents a single file rename operation in the plan
type PlannedOperation struct {
	ID          string           `json:"id"`
	VideoFile   models.VideoFile `json:"video_file"`
	TargetPath  string           `json:"target_path"`
	TargetName  string           `json:"target_name"`
	MediaInfo   interface{}      `json:"media_info,omitempty"`
	Status      OperationStatus  `json:"status"`
	Error       string           `json:"error,omitempty"`
	ConflictIDs []string         `json:"conflict_ids,omitempty"`
}

// Conflict represents a naming conflict between operations
type Conflict struct {
	ID           string              `json:"id"`
	TargetPath   string              `json:"target_path"`
	OperationIDs []string            `json:"operation_ids"`
	ConflictType models.ConflictType `json:"conflict_type"`
	Resolved     bool                `json:"resolved"`
	Resolution   ConflictResolution  `json:"resolution,omitempty"`
}

// ConflictResolution represents how a conflict was resolved
type ConflictResolution struct {
	Strategy      string            `json:"strategy"`
	Modifications map[string]string `json:"modifications"` // operationID -> new target path
	Timestamp     time.Time         `json:"timestamp"`
}

// PlanSummary provides an overview of the plan
type PlanSummary struct {
	TotalOperations      int `json:"total_operations"`
	ReadyOperations      int `json:"ready_operations"`
	ConflictedOperations int `json:"conflicted_operations"`
	SkippedOperations    int `json:"skipped_operations"`
	ErrorOperations      int `json:"error_operations"`
	TotalConflicts       int `json:"total_conflicts"`
	ResolvedConflicts    int `json:"resolved_conflicts"`
}

// OperationStatus represents the status of a planned operation
type OperationStatus string

const (
	OperationStatusPending    OperationStatus = "pending"
	OperationStatusReady      OperationStatus = "ready"
	OperationStatusConflicted OperationStatus = "conflicted"
	OperationStatusSkipped    OperationStatus = "skipped"
	OperationStatusError      OperationStatus = "error"
)
