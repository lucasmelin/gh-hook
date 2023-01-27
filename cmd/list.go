package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.Flags().StringVarP(&repoOverride, "repo", "", "", "Specify a repository. If omitted, uses current repository")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all GitHub webhooks for a repository",
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

		currentHooks, err := getWebhooks(repo)
		if err != nil {
			return fmt.Errorf("could not get webhooks: %w\n", err)
		}
		choices := formatHookChoices(currentHooks)
		if len(choices) == 0 {
			fmt.Printf("%s/%s has no webhooks\n", repo.Owner(), repo.Name())
			return nil
		}
		for _, hook := range choices {
			fmt.Println(hook)
		}

		return nil
	},
}

func formatHookChoices(currentHooks []Hook) []string {
	var choices []string
	for _, choice := range currentHooks {
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
	return choices
}

func getWebhooks(repo repository.Repository) ([]Hook, error) {
	hookOpts := api.ClientOptions{
		Host: repo.Host(),
	}
	client, err := gh.RESTClient(&hookOpts)
	if err != nil {
		return nil, err
	}
	response := []Hook{}
	apiUrl := fmt.Sprintf("repos/%s/%s/hooks", repo.Owner(), repo.Name())
	err = client.Get(apiUrl, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
