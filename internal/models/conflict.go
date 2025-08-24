package models

// ConflictType represents the type of conflict
type ConflictType string

const (
	ConflictTypeTargetExists   ConflictType = "target_exists"   // Target file already exists on disk
	ConflictTypeMultipleSource ConflictType = "multiple_source" // Multiple source files want same target
)

// ConflictStrategy defines how to handle naming conflicts
type ConflictStrategy int

const (
	SkipConflict ConflictStrategy = iota
	AppendNumber
	AppendTimestamp
	PromptUser
	Overwrite
)
