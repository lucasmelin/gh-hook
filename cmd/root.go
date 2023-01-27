package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hook",
	Short: "Hook makes it easy to manage your repository webhooks.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() {
	rootCmd.PersistentFlags().String("repo", "", "Specify a repository. If omitted, uses the current repository.")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
