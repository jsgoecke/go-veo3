package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "veo3",
	Short: "Veo3 CLI for Google's Video Generation API",
	Long: `Veo3 CLI is a command-line utility for interacting with Google's Veo 3.1 model.
It supports text-to-video, image-to-video, frame interpolation, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
