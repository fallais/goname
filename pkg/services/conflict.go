package services

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goname/internal/models"
	"goname/pkg/log"

	"go.uber.org/zap"
)

// ConflictResult represents the result of conflict resolution
type ConflictResult struct {
	ResolvedPath string
	Action       string
	Skipped      bool
}

// ConflictResolver handles file naming conflicts during rename operations
type ConflictResolver struct {
	strategy    models.ConflictStrategy
	interactive bool
}

// NewConflictResolver creates a new conflict resolver with the specified strategy
func NewConflictResolver(strategy models.ConflictStrategy) *ConflictResolver {
	return &ConflictResolver{
		strategy:    strategy,
		interactive: strategy == models.PromptUser,
	}
}

// SetInteractive enables or disables interactive mode for prompts
func (cr *ConflictResolver) SetInteractive(interactive bool) {
	cr.interactive = interactive
}

// CheckConflict checks if a target path already exists
func (cr *ConflictResolver) CheckConflict(targetPath string) bool {
	if _, err := os.Stat(targetPath); err != nil {
		return os.IsExist(err)
	}
	return true
}

// ResolveConflict resolves a naming conflict and returns the final path to use
func (cr *ConflictResolver) ResolveConflict(targetPath string) (*ConflictResult, error) {
	result := &ConflictResult{
		ResolvedPath: targetPath,
		Action:       "none",
		Skipped:      false,
	}

	// If no conflict exists, return the original path
	if !cr.CheckConflict(targetPath) {
		return result, nil
	}

	log.Info("Conflict detected", zap.String("target_path", targetPath))

	switch cr.strategy {
	case models.SkipConflict:
		result.Skipped = true
		result.Action = "skipped"
		result.ResolvedPath = ""
		return result, nil

	case models.AppendNumber:
		resolvedPath, err := cr.appendNumber(targetPath)
		if err != nil {
			return nil, err
		}
		result.ResolvedPath = resolvedPath
		result.Action = "append_number"
		return result, nil

	case models.AppendTimestamp:
		resolvedPath, err := cr.appendTimestamp(targetPath)
		if err != nil {
			return nil, err
		}
		result.ResolvedPath = resolvedPath
		result.Action = "append_timestamp"
		return result, nil

	case models.PromptUser:
		if !cr.interactive {
			// Fall back to skip if not interactive
			result.Skipped = true
			result.Action = "skipped"
			result.ResolvedPath = ""
			return result, nil
		}
		resolvedPath, skipped, err := cr.promptUser(targetPath)
		if err != nil {
			return nil, err
		}
		result.ResolvedPath = resolvedPath
		result.Skipped = skipped
		result.Action = "user_choice"
		return result, nil

	case models.Overwrite:
		result.Action = "overwrite"
		return result, nil

	default:
		return nil, fmt.Errorf("unknown conflict resolution strategy: %d", cr.strategy)
	}
}

// appendNumber adds a number suffix to avoid conflicts
func (cr *ConflictResolver) appendNumber(targetPath string) (string, error) {
	ext := filepath.Ext(targetPath)
	base := strings.TrimSuffix(targetPath, ext)

	for i := 1; i <= 999; i++ { // Prevent infinite loop
		newPath := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath, nil
		}
	}

	return "", fmt.Errorf("unable to resolve conflict for %s after 999 attempts", targetPath)
}

// appendTimestamp adds a timestamp suffix to avoid conflicts
func (cr *ConflictResolver) appendTimestamp(targetPath string) (string, error) {
	ext := filepath.Ext(targetPath)
	base := strings.TrimSuffix(targetPath, ext)
	timestamp := time.Now().Format("20060102_150405")

	newPath := fmt.Sprintf("%s_%s%s", base, timestamp, ext)
	return newPath, nil
}

// promptUser interactively asks the user how to handle the conflict
func (cr *ConflictResolver) promptUser(targetPath string) (string, bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nFile already exists: %s\n", filepath.Base(targetPath))
	fmt.Println("Choose an action:")
	fmt.Println("  [o] Overwrite existing file")
	fmt.Println("  [s] Skip this file")
	fmt.Println("  [a] Append number (e.g., filename (1).ext)")
	fmt.Println("  [t] Append timestamp")
	fmt.Print("Your choice [o/s/a/t]: ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return "", false, err
	}

	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "o", "overwrite":
		return targetPath, false, nil
	case "s", "skip":
		return "", true, nil
	case "a", "append":
		resolvedPath, err := cr.appendNumber(targetPath)
		if err != nil {
			return "", false, err
		}
		return resolvedPath, false, nil
	case "t", "timestamp":
		resolvedPath, err := cr.appendTimestamp(targetPath)
		if err != nil {
			return "", false, err
		}
		return resolvedPath, false, nil
	default:
		fmt.Println("Invalid choice. Skipping file.")
		return "", true, nil
	}
}

// GetStrategyName returns a human-readable name for the conflict strategy
func (cr *ConflictResolver) GetStrategyName() string {
	switch cr.strategy {
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

// ParseConflictStrategy parses a string into a ConflictStrategy
func ParseConflictStrategy(strategy string) (models.ConflictStrategy, error) {
	switch strings.ToLower(strategy) {
	case "skip":
		return models.SkipConflict, nil
	case "append":
		return models.AppendNumber, nil
	case "timestamp":
		return models.AppendTimestamp, nil
	case "prompt":
		return models.PromptUser, nil
	case "overwrite":
		return models.Overwrite, nil
	default:
		return models.AppendNumber, fmt.Errorf("unknown conflict strategy: %s", strategy)
	}
}
