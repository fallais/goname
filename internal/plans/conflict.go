package plans

import (
	"time"

	"goname/internal/models"
)

// Conflict represents a naming conflict between changes
type Conflict struct {
	ID           string              `json:"id"`
	TargetPath   string              `json:"target_path"`
	ChangeIDs    []string            `json:"change_ids"`
	ConflictType models.ConflictType `json:"conflict_type"`
	Resolved     bool                `json:"resolved"`
	Resolution   ConflictResolution  `json:"resolution,omitempty"`
}

// ConflictResolution represents how a conflict was resolved
type ConflictResolution struct {
	Strategy      string            `json:"strategy"`
	Modifications map[string]string `json:"modifications"` // changeID -> new target path
	Timestamp     time.Time         `json:"timestamp"`
}

// removeConflictID removes a specific conflict ID from a slice of conflict IDs
func removeConflictID(conflictIDs []string, conflictID string) []string {
	result := make([]string, 0, len(conflictIDs))
	for _, id := range conflictIDs {
		if id != conflictID {
			result = append(result, id)
		}
	}
	return result
}
