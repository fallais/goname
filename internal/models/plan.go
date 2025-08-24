package models

import "time"

// Plan represents a complete rename plan with all operations and conflicts
type Plan struct {
	ID         string             `json:"id"`
	Timestamp  time.Time          `json:"timestamp"`
	Operations []PlannedOperation `json:"operations"`
	Conflicts  []Conflict         `json:"conflicts"`
	Resolved   bool               `json:"resolved"`
	Summary    PlanSummary        `json:"summary"`
}

// PlannedOperation represents a single file rename operation in the plan
type PlannedOperation struct {
	ID          string          `json:"id"`
	VideoFile   VideoFile       `json:"video_file"`
	TargetPath  string          `json:"target_path"`
	TargetName  string          `json:"target_name"`
	MediaInfo   interface{}     `json:"media_info,omitempty"`
	Status      OperationStatus `json:"status"`
	Error       string          `json:"error,omitempty"`
	ConflictIDs []string        `json:"conflict_ids,omitempty"`
}

// Conflict represents a naming conflict between operations
type Conflict struct {
	ID           string             `json:"id"`
	TargetPath   string             `json:"target_path"`
	OperationIDs []string           `json:"operation_ids"`
	ConflictType ConflictType       `json:"conflict_type"`
	Resolved     bool               `json:"resolved"`
	Resolution   ConflictResolution `json:"resolution,omitempty"`
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
