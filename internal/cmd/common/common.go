package common

import (
	"fmt"
	"goname/internal/models"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

var (
	Green  = color.New(color.FgGreen, color.Bold)
	Red    = color.New(color.FgRed, color.Bold)
	Yellow = color.New(color.FgYellow)
	Blue   = color.New(color.FgBlue)
	Cyan   = color.New(color.FgCyan)
	Gray   = color.New(color.FgHiBlack)
)

// DisplayPlanResults displays the results of a rename plan
func DisplayPlanResults(plan *models.Plan) {
	alreadyCorrectCount := 0
	needsRenameCount := 0
	errorCount := 0
	skippedCount := 0

	fmt.Println("GoName will perform the following actions:")
	fmt.Println()

	for _, operation := range plan.Operations {
		switch operation.Status {
		case models.OperationStatusReady:
			// Check if the current filename already matches the proposed new filename
			currentBaseName := strings.TrimSuffix(operation.VideoFile.OriginalName, filepath.Ext(operation.VideoFile.OriginalName))
			proposedBaseName := strings.TrimSuffix(operation.TargetName, filepath.Ext(operation.TargetName))

			if currentBaseName == proposedBaseName {
				// File is already correctly named
				alreadyCorrectCount++
				Green.Printf("%s\n", operation.VideoFile.OriginalName)
			} else {
				// File needs to be renamed
				needsRenameCount++
				fmt.Printf("%s → %s\n", operation.VideoFile.OriginalName, Yellow.Sprint(operation.TargetName))
			}

		case models.OperationStatusSkipped:
			skippedCount++
			fmt.Printf("%s: %s\n", operation.VideoFile.OriginalName, Blue.Sprint("SKIPPED"))
			if operation.Error != "" {
				fmt.Printf("    Reason: %s\n", operation.Error)
			}

		case models.OperationStatusError:
			errorCount++
			fmt.Printf("%s: %v\n", operation.VideoFile.OriginalName, Red.Sprint(operation.Error))

		case models.OperationStatusConflicted:
			// This should not happen after conflict resolution
			errorCount++
			fmt.Printf("%s: %s\n", operation.VideoFile.OriginalName, Red.Sprint("UNRESOLVED CONFLICT"))
		}
	}

	// Summary
	printPlanSummary(plan, alreadyCorrectCount, needsRenameCount, errorCount, skippedCount)
}

// printPlanSummary prints a summary of the plan results
func printPlanSummary(plan *models.Plan, alreadyCorrectCount, needsRenameCount, errorCount, skippedCount int) {
	fmt.Println()
	fmt.Println(color.HiBlackString("─────────────────────────────────────────────────────────────"))
	fmt.Printf("Plan Summary: ")
	Green.Printf("%d correct", alreadyCorrectCount)
	fmt.Print(", ")
	Yellow.Printf("%d to rename", needsRenameCount)
	fmt.Print(", ")
	if skippedCount > 0 {
		Blue.Printf("%d skipped", skippedCount)
		fmt.Print(", ")
	}
	if errorCount > 0 {
		Red.Printf("%d errors", errorCount)
	} else {
		fmt.Print("0 errors")
	}
	fmt.Printf(", %d total\n", len(plan.Operations))

	if len(plan.Conflicts) > 0 {
		fmt.Printf("Conflicts: %d total, %d resolved\n", len(plan.Conflicts), plan.Summary.ResolvedConflicts)
	}

	if needsRenameCount > 0 {
		fmt.Println()
		Yellow.Println("To apply these changes, run: goname apply")
	}
}
