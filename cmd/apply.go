package cmd

import (
	"github.com/spf13/cobra"

	"goname/internal/cmd/apply"
)

var (
	applyInputDir    string
	applyRecursive   bool
	applyDryRun      bool
	applyTmdbAPIKey  string
	applyMediaType   string
	applyInteractive bool
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

	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "Show what would be renamed without actually renaming")
	applyCmd.Flags().StringVarP(&applyMediaType, "type", "t", "auto", "Media type: movie, tv, or auto")
	applyCmd.Flags().BoolVarP(&applyInteractive, "interactive", "i", false, "Interactive mode for manual confirmation")
}
