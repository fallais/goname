package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long: `Manage GoName configuration settings.

You can set, get, and list configuration values using this command.
Configuration values are stored in ~/.goname.yaml by default.

Examples:
  # Set TMDB API key
  goname config set tmdb.api_key YOUR_API_KEY
  
  # Get current API key
  goname config get tmdb.api_key
  
  # List all configuration
  goname config list
  
  # Initialize default configuration file
  goname config init`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		viper.Set(key, value)
		if err := viper.WriteConfig(); err != nil {
			// If config file doesn't exist, create it
			if err := viper.SafeWriteConfig(); err != nil {
				fmt.Printf("Error writing config: %v\n", err)
				return
			}
		}

		fmt.Printf("Set %s = %s\n", key, value)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := viper.Get(key)

		if value == nil {
			fmt.Printf("%s is not set\n", key)
		} else {
			fmt.Printf("%s = %v\n", key, value)
		}
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		settings := viper.AllSettings()

		if len(settings) == 0 {
			fmt.Println("No configuration values set")
			return
		}

		fmt.Println("Configuration:")
		for key, value := range settings {
			fmt.Printf("  %s = %v\n", key, value)
		}
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}

		configFile := filepath.Join(home, ".goname.yaml")

		// Check if config file already exists
		if _, err := os.Stat(configFile); err == nil {
			fmt.Printf("Configuration file already exists at %s\n", configFile)
			return
		}

		// Create default config
		defaultConfig := `# GoName Configuration File
# Set your TMDB API key here
tmdb:
  api_key: ""

# Default settings
rename:
  interactive: false
  dry_run: false
  recursive: false
  
# Output formatting
format:
  movie: "{title} ({year})"
  tv: "{show} - S{season:02d}E{episode:02d} - {episode_name}"
`

		if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
			fmt.Printf("Error creating config file: %v\n", err)
			return
		}

		fmt.Printf("Created configuration file at %s\n", configFile)
		fmt.Println("Edit this file to set your TMDB API key and preferences.")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configInitCmd)
}
