package plans

type Action rune

const (
	ActionNoop   Action = 0
	ActionRename Action = '~'
	ActionSkip   Action = '-'
)

type FileInfo struct {
	Path     string `json:"path"`
	FileName string `json:"file_name"`
}

// Change represents a proposed change to a video file.
type Change struct {
	ID string `json:"id"`

	// Action to be performed
	Action Action `json:"action"`

	Before FileInfo `json:"before"`
	After  FileInfo `json:"after"`

	// ConflictIDs tracks which conflicts affect this change
	ConflictIDs []string `json:"conflict_ids,omitempty"`

	// Error stores any error that occurred during planning
	Error string `json:"error,omitempty"`
}

// IsConflicting returns true if this change has unresolved conflicts
func (c *Change) IsConflicting() bool {
	return len(c.ConflictIDs) > 0
}
