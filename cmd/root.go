package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "bytes-be",
	Short: "bytes backend server",
	Long:  "bytes backend server",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
