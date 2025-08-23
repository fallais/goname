package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goname",
	Short: "A tool that helps you rename your video files",
	Long: `GoName is a CLI application that helps you rename video files by fetching
information from TheMovieDB, TheTVDB, and other databases.

It can rename large amounts of video files automatically by matching them
with movie or TV show information.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goname.yaml)")
	rootCmd.PersistentFlags().StringVarP(&applyInputDir, "dir", "d", ".", "Directory to scan for video files")
	rootCmd.PersistentFlags().BoolVarP(&applyRecursive, "recursive", "r", false, "Scan directories recursively")
	rootCmd.PersistentFlags().StringVar(&applyTmdbAPIKey, "tmdb_api_key", "", "TMDB API key (can also be set via TMDB_API_KEY env var)")
	rootCmd.PersistentFlags().StringVarP(&planMediaType, "type", "t", "auto", "Media type: movie, tv, or auto")

	// Bind flags to viper
	viper.BindPFlag("tmdb.api_key", rootCmd.Flags().Lookup("tmdb_api_key"))
	viper.BindPFlag("media.type", rootCmd.Flags().Lookup("type"))

	// Env
	viper.BindEnv("tmdb.api_key", "TMDB_API_KEY")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".goname" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".goname")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	viper.ReadInConfig()
}
