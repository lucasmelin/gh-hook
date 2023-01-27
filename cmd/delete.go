package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/lucasmelin/gh-hook/tui"
	"github.com/spf13/cobra"
)

type Hook struct {
	Id     int      `json:"id,omitempty"`
	Active bool     `json:"active,omitempty"`
	Events []string `json:"events,omitempty"`
	Config HookConfig
}

func init() {
	deleteCmd.Flags().StringVarP(&repoOverride, "repo", "", "", "Specify a repository. If omitted, uses current repository")
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a GitHub webhook",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var repo repository.Repository
		var err error

		if repoOverride != "" {
			repo, err = repository.Parse(repoOverride)
		} else {
			repo, err = gh.CurrentRepository()
		}
		if err != nil {
			return fmt.Errorf("could not determine the repo to use: %w\n", err)
		}

		hookOpts := api.ClientOptions{
			Host: repo.Host(),
		}
		client, err := gh.RESTClient(&hookOpts)
		if err != nil {
			log.Fatal(err)
		}
		response := []Hook{}
		apiUrl := fmt.Sprintf("repos/%s/%s/hooks", repo.Owner(), repo.Name())
		err = client.Get(apiUrl, &response)
		if err != nil {
			log.Fatal(err)
		}
		var choices []string
		for _, choice := range response {
			var displayText string
			if choice.Active {
				displayText += "✓ "
			} else {
				displayText += "• "
			}
			stringEvents := strings.Join(choice.Events, ", ")
			if len(stringEvents) > 23 {
				stringEvents = stringEvents[:23] + "…"
			}
			stringEvents = "(" + stringEvents + ")"

			displayText += strconv.Itoa(choice.Id) + " - " + choice.Config.Url + " " + stringEvents
			choices = append(choices, displayText)
		}

		hooksToDelete, err := tui.Choose("Which webhooks would you like to delete?", choices, 0)
		var deleteIds []string
		for _, hook := range hooksToDelete {
			_, withoutPrefix, _ := strings.Cut(hook, " ")
			id, _, _ := strings.Cut(withoutPrefix, " ")
			deleteIds = append(deleteIds, id)
		}

		for _, hookId := range deleteIds {
			fmt.Printf("Deleting %s\n", hookId)
			apiUrl := fmt.Sprintf("repos/%s/%s/hooks/%s", repo.Owner(), repo.Name(), hookId)
			err = client.Delete(apiUrl, nil)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	},
}
