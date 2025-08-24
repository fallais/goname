package cmd

import (
	"goname/internal/cmd/plan"

	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show what files would be renamed without actually renaming them",
	Long: `Plan shows what video files would be renamed by fetching information from TMDB
and displaying the proposed changes. This is similar to 'terraform plan' - it shows
what would happen without making any actual changes.

The command will scan for video files, attempt to match them with movies or TV shows
from TMDB, and display the current filename alongside the proposed new filename.`,

	Run: plan.Run,
}

func init() {
	rootCmd.AddCommand(planCmd)
}
