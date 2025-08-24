package cmd

import (
	"goname/internal/cmd/state/ls"

	"github.com/spf13/cobra"
)

var (
	stateLsActive bool
	stateLsLimit  int
)

// stateLsCmd represents the state ls command
var stateLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List rename operations from state",
	Long: `List rename operations that have been performed by GoName.

This command shows a history of all rename operations, including:
- Original and new file names
- Timestamp when the operation was performed  
- Whether the operation has been reverted
- Media information that was used for the rename

Examples:
  # List all rename operations
  goname state ls
  
  # List only active (non-reverted) operations
  goname state ls --active
  
  # List only the last 10 operations
  goname state ls --limit 10`,
	Run: ls.Run,
}

func init() {
	stateLsCmd.Flags().BoolVar(&stateLsActive, "active", false, "Show only active (non-reverted) operations")
	stateLsCmd.Flags().IntVarP(&stateLsLimit, "limit", "l", 0, "Limit the number of operations to show (0 for all)")
}
