package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a GitHub webhook",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("This is the delete command")
	},
}
