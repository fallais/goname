package ls

import (
	"fmt"
	"sort"

	"goname/internal/models"
	"goname/pkg/log"
	"goname/pkg/services"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Run(cmd *cobra.Command, args []string) {
	log.Debug("goname state ls is starting", zap.String("command", "state ls"))

	stateService, err := services.NewStateService()
	if err != nil {
		log.Fatal("failed to initialize state service", zap.Error(err))
	}

	state, err := stateService.LoadState()
	if err != nil {
		log.Fatal("failed to load state", zap.Error(err))
	}

	// Filter entries based on flags
	entries := state.Entries
	activeOnly, _ := cmd.Flags().GetBool("active")
	if activeOnly {
		var filteredEntries []models.StateEntry
		for _, entry := range entries {
			if !entry.Reverted {
				filteredEntries = append(filteredEntries, entry)
			}
		}
		entries = filteredEntries
	}

	// Sort entries by timestamp (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	// Apply limit if specified
	limit, _ := cmd.Flags().GetInt("limit")
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	if len(entries) == 0 {
		color.Yellow("No rename operations found.")
		return
	}

	// Color setup
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	gray := color.New(color.FgHiBlack)

	fmt.Printf("Found %d rename operation(s):\n\n", len(entries))

	for _, entry := range entries {
		// Status indicator
		if entry.Reverted {
			red.Print("✗ REVERTED ")
		} else {
			green.Print("✓ ACTIVE   ")
		}

		// Entry ID and timestamp
		cyan.Printf("[%s] ", entry.ID[:8])
		gray.Printf("%s\n", entry.Timestamp.Format("2006-01-02 15:04:05"))

		// Original and new names
		fmt.Printf("  Original: %s\n", entry.OriginalName)
		fmt.Printf("  New:      %s\n", entry.NewName)

		// Media info if available
		if entry.MediaInfo != nil {
			switch mediaInfo := entry.MediaInfo.(type) {
			case map[string]interface{}:
				if title, ok := mediaInfo["title"].(string); ok {
					yellow.Printf("  Media:    %s", title)
					if year, ok := mediaInfo["release_date"].(string); ok && len(year) >= 4 {
						yellow.Printf(" (%s)", year[:4])
					}
					fmt.Println()
				} else if name, ok := mediaInfo["name"].(string); ok {
					yellow.Printf("  Media:    %s", name)
					if year, ok := mediaInfo["first_air_date"].(string); ok && len(year) >= 4 {
						yellow.Printf(" (%s)", year[:4])
					}
					fmt.Println()
				}
			}
		}

		fmt.Println()
	}

	// Summary
	if activeOnly {
		fmt.Printf("Showing %d active operations", len(entries))
	} else {
		activeCount := 0
		for _, entry := range state.Entries {
			if !entry.Reverted {
				activeCount++
			}
		}
		fmt.Printf("Showing %d operations (%d active, %d reverted)",
			len(entries), activeCount, len(state.Entries)-activeCount)
	}

	if limit > 0 && len(state.Entries) > limit {
		fmt.Printf(" (limited to %d)", limit)
	}
	fmt.Println()
}
