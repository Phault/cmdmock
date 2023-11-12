package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cmdmock",
	Short: "",
	Long:  ``,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var OutputDir string

func init() {
	rootCmd.PersistentFlags().StringVarP(&OutputDir, "outputDir", "o", "./", "The directory to store recordings, relative to working directory or an absolute path.")
}
