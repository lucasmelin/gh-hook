package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/lucasmelin/gh-hook/tui"
	"github.com/spf13/cobra"
)

func NewCmdCreate() *cobra.Command {
	var createCmd = &cobra.Command{
		Use:          "create",
		Short:        "Create a new repository webhook",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := getRepo(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("Creating new webhook for %s\n", repo.Name())

			refreshEvents, _ := cmd.Flags().GetBool("refresh-events")
			events, err := getEvents(refreshEvents)
			if err != nil {
				return fmt.Errorf("could not get events: %w\n", err)
			}

			fileInput, _ := cmd.Flags().GetString("file")

			var newHook Hook
			if len(fileInput) > 0 {
				file, err := os.Open(fileInput)
				if err != nil {
					return fmt.Errorf("could not open JSON file: %w\n", err)
				}
				newHook, err = hookFromInput(file)
			} else {
				newHook, err = hookFromPrompt(events)
				if err != nil {
					return err
				}
			}

			if err := createHook(repo, newHook); err != nil {
				return err
			}
			fmt.Println("Successfully created hook 🪝")
			return nil
		},
	}
	createCmd.Flags().Bool("refresh-events", false, "Download the list of events from https://octokit.github.io/webhooks By default, a hardcoded list of known events will be used.")
	createCmd.Flags().String("file", "", "Provide the webhook data as a JSON file.")
	return createCmd
}

func hookFromInput(file io.Reader) (Hook, error) {
	newHook := Hook{}
	parser := json.NewDecoder(file)
	if err := parser.Decode(&newHook); err != nil {
		return newHook, fmt.Errorf("could not parse JSON data %+v: %w", file, err)
	}
	return newHook, nil
}

func hookFromPrompt(events []string) (Hook, error) {
	hookUrl, err := tui.Input(false, "Webhook URL: ")
	if err != nil {
		return Hook{}, fmt.Errorf("could not get webhook URL: %w\n", err)
	}
	hookEvents, err := tui.ChooseMany("Events to receive", events)
	if err != nil {
		return Hook{}, fmt.Errorf("could not choose events: %w\n", err)
	}
	secret, err := tui.Input(true, "Webhook secret (optional): ")
	if err != nil {
		return Hook{}, fmt.Errorf("could not get webhook secret: %w\n", err)
	}
	contentType, err := tui.ChooseOne("Content Type", []string{"json", "form"})
	if err != nil {
		return Hook{}, fmt.Errorf("could not choose content type: %w\n", err)
	}
	sslChoice, err := tui.ChooseOne("Insecure SSL", []string{"true", "false"})
	if err != nil {
		return Hook{}, fmt.Errorf("could not choose insecure SSL option: %w\n", err)
	}
	ssl := "0"
	if sslChoice == "true" {
		ssl = "1"
	}
	activeChoice, err := tui.ChooseOne("Webhook Active", []string{"true", "false"})
	if err != nil {
		return Hook{}, fmt.Errorf("could not choose webhook active option: %w\n", err)
	}
	active := false
	if activeChoice == "true" {
		active = true
	}

	return Hook{
		Name:   "web",
		Active: active,
		Events: hookEvents,
		Config: HookConfig{
			Url:         hookUrl,
			ContentType: contentType,
			InsecureSSL: ssl,
			Secret:      secret,
		},
	}, nil
}

func createHook(repo repository.Repository, data Hook) error {
	hookOpts := api.ClientOptions{
		Host: repo.Host(),
	}
	client, err := gh.RESTClient(&hookOpts)
	if err != nil {
		return fmt.Errorf("error creating REST client: %w\n", err)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("could not convert responses to JSON: %w\n", err)
	}

	apiUrl := fmt.Sprintf("repos/%s/%s/hooks", repo.Owner(), repo.Name())
	if err := client.Post(apiUrl, bytes.NewBuffer(jsonData), nil); err != nil {
		return fmt.Errorf("could not create new webhook: %w\n", err)
	}
	return nil
}

func getEvents(refresh bool) ([]string, error) {
	if !refresh {
		return knownEvents, nil
	}
	client := &http.Client{}
	assetURL := "https://octokit.github.io/webhooks/payload-examples/api.github.com/index.json"
	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, err
	}
	var eventNames []string
	for _, e := range events {
		eventNames = append(eventNames, e.Name)
	}
	return eventNames, nil
}
