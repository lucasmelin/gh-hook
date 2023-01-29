package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/lucasmelin/gh-hook/tui"
	"github.com/spf13/cobra"
)

func init() {
	createCmd.Flags().Bool("refresh-events", false, "Download the list of events from octokit.github.io/webhooks/. By default, a hardcoded list of known events will be used.")
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:          "create",
	Short:        "Create a new repository webhook.",
	Long:         ``,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := getRepo(cmd)
		if err != nil {
			return err
		}

		fmt.Printf("Creating new webhook for %s\n", repo.Name())

		hookUrl, err := tui.Input(false, "Webhook URL: ")
		if err != nil {
			return fmt.Errorf("error getting webhook URL: %w\n", err)
		}

		refreshEvents, _ := cmd.Flags().GetBool("refresh-events")
		events, err := getEvents(refreshEvents)
		if err != nil {
			return fmt.Errorf("error getting events: %w\n", err)
		}

		hookEvents, err := tui.Choose("Events to receive", events, 0)
		if err != nil {
			return fmt.Errorf("error choosing events: %w\n", err)
		}
		secret, err := tui.Input(true, "Webhook secret (optional): ")
		if err != nil {
			return fmt.Errorf("error entering webhook secret: %w\n", err)
		}
		contentTypeChoice, err := tui.Choose("Content Type", []string{"json", "form"}, 1)
		if err != nil {
			return fmt.Errorf("error choosing content type: %w\n", err)
		}
		if len(contentTypeChoice) == 0 {
			return fmt.Errorf("no content type selected")
		}
		contentType := contentTypeChoice[0]
		sslChoice, err := tui.Choose("Insecure SSL", []string{"true", "false"}, 1)
		if err != nil {
			return fmt.Errorf("error choosing insecure SSL option: %w\n", err)
		}
		if len(sslChoice) == 0 {
			return fmt.Errorf("no SSL choice selected")
		}
		var ssl string
		if sslChoice[0] == "true" {
			ssl = "1"
		} else {
			ssl = "0"
		}
		activeChoice, err := tui.Choose("Webhook Active", []string{"true", "false"}, 1)
		if err != nil {
			return fmt.Errorf("error choosing webhook active option: %w\n", err)
		}
		if len(activeChoice) == 0 {
			return fmt.Errorf("no active choice selected")
		}
		active := false
		if activeChoice[0] == "true" {
			active = true
		}

		value := Hook{
			Name:   "web",
			Active: active,
			Events: hookEvents,
			Config: HookConfig{
				Url:         hookUrl,
				ContentType: contentType,
				InsecureSSL: ssl,
				Secret:      secret,
			},
		}

		if err := createHook(repo, value); err != nil {
			return err
		}
		fmt.Println("Successfully created hook 🪝")
		return nil
	},
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
		return fmt.Errorf("error converting resoponses to JSON: %w\n", err)
	}
	apiUrl := fmt.Sprintf("repos/%s/%s/hooks", repo.Owner(), repo.Name())
	err = client.Post(apiUrl, bytes.NewBuffer(jsonData), nil)
	if err != nil {
		return fmt.Errorf("error creating new webhook: %w\n", err)
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
