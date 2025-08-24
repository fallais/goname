package cmd

import (
	"github.com/spf13/cobra"

	"goname/internal/cmd/apply"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the planned renames to video files using TMDB data",
	Long: `Apply renames video files in a directory by fetching information from The Movie Database (TMDB).

This command is similar to 'terraform apply' - it performs the actual renaming of files
based on TMDB lookups. Use 'goname plan' first to preview what changes will be made.

Examples:
  # Apply renames for movies in current directory
  goname apply --type movie --api-key YOUR_API_KEY
  
  # Apply renames for TV shows recursively with dry-run
  goname apply --dir /path/to/shows --type tv --recursive --dry-run --api-key YOUR_API_KEY
  
  # Interactive mode for manual confirmation
  goname apply --interactive --api-key YOUR_API_KEY`,

	Run: apply.Run,
}

func init() {
	rootCmd.AddCommand(applyCmd)

	applyCmd.Flags().Bool("auto-approve", false, "Will not prompt for confirmation before applying changes")
	applyCmd.Flags().BoolP("interactive", "i", false, "Interactive mode for manual confirmation")
}
