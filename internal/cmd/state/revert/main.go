package revert

import (
	"fmt"
	"os"
	"path/filepath"

	"goname/internal/models"
	"goname/pkg/log"
	"goname/pkg/services"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Run(cmd *cobra.Command, args []string) {
	log.Debug("goname state revert is starting", zap.String("command", "state revert"))

	stateService, err := services.NewStateService()
	if err != nil {
		log.Fatal("failed to initialize state service", zap.Error(err))
	}

	fileService := services.NewFileService()

	// Get flags
	id, _ := cmd.Flags().GetString("id")
	last, _ := cmd.Flags().GetBool("last")
	all, _ := cmd.Flags().GetBool("all")

	// Validate flags
	flagCount := 0
	if id != "" {
		flagCount++
	}
	if last {
		flagCount++
	}
	if all {
		flagCount++
	}

	if flagCount == 0 {
		color.Red("Error: You must specify one of --id, --last, or --all")
		return
	}

	if flagCount > 1 {
		color.Red("Error: You can only specify one of --id, --last, or --all")
		return
	}

	// Color setup
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)

	var entriesToRevert []models.StateEntry

	if id != "" {
		// Revert specific entry by ID
		entry, err := stateService.GetEntryByID(id)
		if err != nil {
			red.Printf("Error: %v\n", err)
			return
		}

		if entry.Reverted {
			yellow.Printf("Entry %s is already reverted\n", id)
			return
		}

		entriesToRevert = []models.StateEntry{*entry}
	} else if last {
		// Revert last active entry
		entry, err := stateService.GetLastActiveEntry()
		if err != nil {
			red.Printf("Error: %v\n", err)
			return
		}

		entriesToRevert = []models.StateEntry{*entry}
	} else if all {
		// Revert all active entries
		entries, err := stateService.GetActiveEntries()
		if err != nil {
			red.Printf("Error: %v\n", err)
			return
		}

		if len(entries) == 0 {
			yellow.Println("No active entries to revert")
			return
		}

		entriesToRevert = entries

		// Confirmation for reverting all
		fmt.Printf("This will revert %d operations. Are you sure? (y/N): ", len(entries))
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation cancelled")
			return
		}
	}

	// Perform the reverts
	successCount := 0
	for _, entry := range entriesToRevert {
		fmt.Printf("Reverting: %s -> %s\n", entry.NewName, entry.OriginalName)

		// Check if the new file still exists
		if _, err := os.Stat(entry.NewPath); os.IsNotExist(err) {
			red.Printf("  ✗ File not found: %s\n", entry.NewPath)
			continue
		}

		// Check if original path would conflict
		originalDir := filepath.Dir(entry.OriginalPath)
		originalPath := filepath.Join(originalDir, entry.OriginalName)

		if _, err := os.Stat(originalPath); err == nil {
			red.Printf("  ✗ Original file already exists: %s\n", originalPath)
			continue
		}

		// Perform the revert
		if err := fileService.RenameFile(entry.NewPath, originalPath); err != nil {
			red.Printf("  ✗ Failed to revert: %v\n", err)
			continue
		}

		// Mark as reverted in state
		if err := stateService.MarkAsReverted(entry.ID); err != nil {
			yellow.Printf("  ! File reverted but failed to update state: %v\n", err)
		} else {
			green.Printf("  ✓ Successfully reverted\n")
			successCount++
		}
	}

	// Summary
	fmt.Printf("\nReverted %d out of %d operations\n", successCount, len(entriesToRevert))

	if successCount < len(entriesToRevert) {
		yellow.Println("Some operations could not be reverted. Check the output above for details.")
	}
}
