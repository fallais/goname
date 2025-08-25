package common

import (
	"fmt"
	"goname/internal/plans"

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
func DisplayPlanResults(plan *plans.Plan) {
	alreadyCorrectCount := 0
	needsRenameCount := 0
	errorCount := 0
	skippedCount := 0

	fmt.Println("GoName will perform the following actions:")
	fmt.Println()

	for _, change := range plan.Changes {
		switch change.Action {
		case plans.ActionRename:
			if change.IsConflicting() {
				// Conflicted change
				needsRenameCount++
				fmt.Printf("%c %s → %s %s\n", change.Action, change.Before.FileName, Yellow.Sprint(change.After.FileName), Red.Sprint("(CONFLICT)"))
			} else if change.Error != "" {
				// Error in change
				errorCount++
				Red.Printf("%c", change.Action)
				fmt.Printf(" %s: %s (%s)\n", change.Before.FileName, Red.Sprint("ERROR"), change.Error)
			} else {
				// Ready to be renamed
				needsRenameCount++
				Yellow.Printf("%c", change.Action)
				fmt.Printf(" %s → %s\n", change.Before.FileName, Yellow.Sprint(change.After.FileName))
			}

		case plans.ActionNoop:
			// File is already correctly named
			alreadyCorrectCount++
			Green.Printf("%c", change.Action)
			fmt.Printf("%s\n", Green.Sprint(change.Before.FileName))

		case plans.ActionSkip:
			skippedCount++
			Blue.Printf("%c", change.Action)
			fmt.Printf(" %s: %s\n", change.Before.FileName, Blue.Sprint("skipped"))
		}
	}

	// Summary
	printPlanSummary(plan, alreadyCorrectCount, needsRenameCount, errorCount, skippedCount)
}

// printPlanSummary prints a summary of the plan results
func printPlanSummary(plan *plans.Plan, alreadyCorrectCount, needsRenameCount, errorCount, skippedCount int) {
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
	fmt.Printf(", %d total\n", len(plan.Changes))

	if len(plan.Conflicts) > 0 {
		fmt.Printf("Conflicts: %d total, %d resolved\n", len(plan.Conflicts), plan.Summary().ResolvedConflicts)
	}

	if needsRenameCount > 0 {
		fmt.Println()
		Yellow.Println("To apply these changes, run: goname apply")
	}
}
