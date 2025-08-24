package models

import "time"

// VideoFile represents a video file to be renamed
type VideoFile struct {
	Path         string
	OriginalName string
	NewName      string
	Size         int64
	ModTime      time.Time
	MediaType    MediaType
}

// RenameResult represents the result of a rename operation
type RenameResult struct {
	VideoFile      VideoFile
	Success        bool
	Error          error
	MediaInfo      interface{} // Can be MovieInfo, TVShowInfo, etc.
	NewFileName    string
	ConflictAction string // Action taken during conflict resolution
}

// StateEntry represents a single rename operation in the state
type StateEntry struct {
	ID           string      `json:"id"`
	Timestamp    time.Time   `json:"timestamp"`
	OriginalPath string      `json:"original_path"`
	NewPath      string      `json:"new_path"`
	OriginalName string      `json:"original_name"`
	NewName      string      `json:"new_name"`
	MediaInfo    interface{} `json:"media_info,omitempty"`
	Reverted     bool        `json:"reverted"`
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

// ConflictType represents the type of conflict
type ConflictType string

const (
	ConflictTypeTargetExists   ConflictType = "target_exists"   // Target file already exists on disk
	ConflictTypeMultipleSource ConflictType = "multiple_source" // Multiple source files want same target
)

// State represents the complete state file
type State struct {
	Version string       `json:"version"`
	Entries []StateEntry `json:"entries"`
}
