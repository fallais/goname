package cmd

import (
	"github.com/spf13/cobra"
)

// stateCmd represents the state command
var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Manage rename operation state",
	Long: `Manage the state of rename operations performed by GoName.

The state command allows you to:
- List previous rename operations
- Revert specific rename operations
- View the history of changes made to your files

Examples:
  # List all rename operations
  goname state ls
  
  # List only unreversed operations
  goname state ls --active
  
  # Revert the last rename operation
  goname state revert
  
  # Revert a specific rename operation
  goname state revert --id abc123`,
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.AddCommand(stateLsCmd)
	stateCmd.AddCommand(stateRevertCmd)
}
