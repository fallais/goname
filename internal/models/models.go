package models

import "time"

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

// State represents the complete state file
type State struct {
	Version string       `json:"version"`
	Entries []StateEntry `json:"entries"`
}
