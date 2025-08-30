package states

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"goru/pkg/log"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// State represents the complete state file
type State struct {
	Version string       `json:"version"`
	Entries []StateEntry `json:"entries"`
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

// StateService handles state file operations
type StateService struct {
	statePath string
}

// NewStateService creates a new StateService
func NewStateService() (*StateService, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	statePath := filepath.Join(home, ".goru", "state.json")

	// Create the directory if it doesn't exist
	stateDir := filepath.Dir(statePath)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	return &StateService{
		statePath: statePath,
	}, nil
}

// LoadState loads the state from file
func (s *StateService) LoadState() (*State, error) {
	if _, err := os.Stat(s.statePath); os.IsNotExist(err) {
		// Return empty state if file doesn't exist
		return &State{
			Version: "1.0",
			Entries: []StateEntry{},
		}, nil
	}

	data, err := os.ReadFile(s.statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// SaveState saves the state to file
func (s *StateService) SaveState(state *State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(s.statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// AddRenameOperation adds a rename operation to the state
func (s *StateService) AddRenameOperation(originalPath, newPath, originalName, newName string, mediaInfo interface{}) error {
	state, err := s.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	entry := StateEntry{
		ID:           uuid.New().String(),
		Timestamp:    time.Now(),
		OriginalPath: originalPath,
		NewPath:      newPath,
		OriginalName: originalName,
		NewName:      newName,
		MediaInfo:    mediaInfo,
		Reverted:     false,
	}

	state.Entries = append(state.Entries, entry)

	if err := s.SaveState(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	log.Debug("Added rename operation to state",
		zap.String("id", entry.ID),
		zap.String("original", originalName),
		zap.String("new", newName))

	return nil
}

// GetActiveEntries returns all non-reverted entries
func (s *StateService) GetActiveEntries() ([]StateEntry, error) {
	state, err := s.LoadState()
	if err != nil {
		return nil, err
	}

	var active []StateEntry
	for _, entry := range state.Entries {
		if !entry.Reverted {
			active = append(active, entry)
		}
	}

	return active, nil
}

// GetEntryByID returns a specific entry by ID
func (s *StateService) GetEntryByID(id string) (*StateEntry, error) {
	state, err := s.LoadState()
	if err != nil {
		return nil, err
	}

	for _, entry := range state.Entries {
		if entry.ID == id {
			return &entry, nil
		}
	}

	return nil, fmt.Errorf("entry with ID %s not found", id)
}

// GetLastActiveEntry returns the most recent non-reverted entry
func (s *StateService) GetLastActiveEntry() (*StateEntry, error) {
	active, err := s.GetActiveEntries()
	if err != nil {
		return nil, err
	}

	if len(active) == 0 {
		return nil, fmt.Errorf("no active entries found")
	}

	// Find the most recent entry
	var lastEntry *StateEntry
	for i := range active {
		if lastEntry == nil || active[i].Timestamp.After(lastEntry.Timestamp) {
			lastEntry = &active[i]
		}
	}

	return lastEntry, nil
}

// MarkAsReverted marks an entry as reverted
func (s *StateService) MarkAsReverted(id string) error {
	state, err := s.LoadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	for i := range state.Entries {
		if state.Entries[i].ID == id {
			state.Entries[i].Reverted = true

			if err := s.SaveState(state); err != nil {
				return fmt.Errorf("failed to save state: %w", err)
			}

			log.Debug("Marked entry as reverted", zap.String("id", id))
			return nil
		}
	}

	return fmt.Errorf("entry with ID %s not found", id)
}
