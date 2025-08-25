package cmd

import (
	"goname/pkg/log"
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
	rootCmd.PersistentFlags().StringP("dir", "d", ".", "Directory to scan for video files")
	rootCmd.PersistentFlags().BoolP("recursive", "r", false, "Scan directories recursively")
	rootCmd.PersistentFlags().String("tmdb_api_key", "", "TMDB API key")
	rootCmd.PersistentFlags().StringP("type", "t", "auto", "Media type: movie, tv, or auto")
	rootCmd.PersistentFlags().String("db", "tmdb", "Database: tmdb (The Movie Database) or tvdb (The TV Database)")
	rootCmd.PersistentFlags().String("conflict", "append", "Conflict resolution strategy: skip, append, timestamp, prompt, overwrite, backup")

	// Bind flags to viper
	viper.BindPFlag("tmdb.api_key", rootCmd.PersistentFlags().Lookup("tmdb_api_key"))
	viper.BindPFlag("type", rootCmd.PersistentFlags().Lookup("type"))
	viper.BindPFlag("db", rootCmd.PersistentFlags().Lookup("db"))
	viper.BindPFlag("conflict", rootCmd.PersistentFlags().Lookup("conflict"))

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

	log.Init(viper.GetBool("debug"))
}
