package cmd

import (
	"fmt"
	"strings"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/lucasmelin/gh-hook/tui"
	"github.com/spf13/cobra"
)

func NewCmdDelete() *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete repository webhooks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := getRepo(cmd)
			if err != nil {
				return err
			}

			response, err := getWebhooks(repo)
			if err != nil {
				return fmt.Errorf("could not get webhooks: %w\n", err)
			}
			choices := formatHookChoices(response)
			if len(choices) == 0 {
				fmt.Printf("%s/%s has no webhooks\n", repo.Owner(), repo.Name())
				return nil
			}

			hooksToDelete, err := tui.ChooseMany("Which webhooks would you like to delete?", choices)
			if err != nil {
				return fmt.Errorf("could not choose webhooks: %w", err)
			}
			var deleteIds []string
			for _, hook := range hooksToDelete {
				_, withoutPrefix, _ := strings.Cut(hook, " ")
				delId, _, _ := strings.Cut(withoutPrefix, " ")
				deleteIds = append(deleteIds, delId)
			}

			return deleteHooks(repo, deleteIds)
		},
	}
	return deleteCmd
}

func deleteHooks(repo repository.Repository, deleteIds []string) error {
	hookOpts := api.ClientOptions{
		Host: repo.Host(),
	}
	client, err := gh.RESTClient(&hookOpts)
	if err != nil {
		return err
	}
	for _, hookId := range deleteIds {
		fmt.Printf("Deleting %s\n", hookId)
		apiUrl := fmt.Sprintf("repos/%s/%s/hooks/%s", repo.Owner(), repo.Name(), hookId)
		err = client.Delete(apiUrl, nil)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Deleted %d hooks 🗑️\n", len(deleteIds))
	return nil
}
