package cmd

import (
	"goname/internal/cmd/state/revert"

	"github.com/spf13/cobra"
)

var (
	stateRevertID   string
	stateRevertLast bool
	stateRevertAll  bool
)

// stateRevertCmd represents the state revert command
var stateRevertCmd = &cobra.Command{
	Use:   "revert",
	Short: "Revert rename operations",
	Long: `Revert rename operations that have been performed by GoName.

This command allows you to undo rename operations by restoring files 
to their original names. You can revert:
- A specific operation by ID
- The last operation performed
- All active operations

Examples:
  # Revert the last rename operation
  goname state revert --last
  
  # Revert a specific operation by ID
  goname state revert --id abc123
  
  # Revert all active operations (be careful!)
  goname state revert --all`,
	Run: revert.Run,
}

func init() {
	stateRevertCmd.Flags().StringVar(&stateRevertID, "id", "", "Revert a specific operation by ID")
	stateRevertCmd.Flags().BoolVar(&stateRevertLast, "last", false, "Revert the last operation")
	stateRevertCmd.Flags().BoolVar(&stateRevertAll, "all", false, "Revert all active operations")
}
