package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/lucasmelin/gh-hook/tui"
	"github.com/spf13/cobra"
)

type Event struct {
	Name string `json:"name"`
}

var repoOverride string

func init() {
	createCmd.Flags().StringVarP(&repoOverride, "repo", "", "", "Specify a repository. If omitted, uses current repository")
	rootCmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
	Use:          "create",
	Short:        "Create a new GitHub webhook",
	Long:         ``,
	SilenceUsage: true,
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

		fmt.Printf("Creating new webhook for %s\n", repo.Name())
		hookUrl, err := tui.Input(false, "Webhook URL: ")
		if err != nil {
			return fmt.Errorf("error entering webhook URL: %w\n", err)
		}

		events := getEvents()
		hookEvents, err := tui.Choose("Events to receive", events, 0)
		if err != nil {
			return fmt.Errorf("error choosing events: %w\n", err)
		}
		secret, err := tui.Input(true, "Webhook secret (optional): ")
		if err != nil {
			return fmt.Errorf("error entering webhook secret: %w\n", err)
		}
		contentType, err := tui.Choose("Content Type", []string{"json", "form"}, 1)
		if err != nil {
			return fmt.Errorf("error choosing content type: %w\n", err)
		}
		sslChoice, err := tui.Choose("Insecure SSL", []string{"true", "false"}, 1)
		if err != nil {
			return fmt.Errorf("error choosing insecure SSL option: %w\n", err)
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
		active := false
		if activeChoice[0] == "true" {
			active = true
		}

		hookOpts := api.ClientOptions{
			Host: repo.Host(),
			Log:  os.Stdout,
		}
		client, err := gh.RESTClient(&hookOpts)
		if err != nil {
			log.Fatal(err)
		}
		value := HookOpts{
			Name:   "web",
			Active: active,
			Events: hookEvents,
			Config: HookConfig{
				Url:         hookUrl,
				ContentType: contentType[0],
				InsecureSSL: ssl,
				Secret:      secret,
			},
		}
		jsonValue, err := json.Marshal(value)
		if err != nil {
			log.Fatal(err)
		}
		apiUrl := fmt.Sprintf("repos/%s/%s/hooks", repo.Owner(), repo.Name())
		err = client.Post(apiUrl, bytes.NewBuffer(jsonValue), nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("successfully created hook")
		return nil
	},
}

func getEvents() []string {
	client := &http.Client{}
	assetURL := "https://octokit.github.io/webhooks/payload-examples/api.github.com/index.json"
	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		log.Fatal(err)
	}
	var eventNames []string
	for _, e := range events {
		eventNames = append(eventNames, e.Name)
	}
	return eventNames
}

type HookOpts struct {
	Name   string     `json:"name,omitempty"`
	Active bool       `json:"active,omitempty"`
	Events []string   `json:"events,omitempty"`
	Config HookConfig `json:"config,omitempty"`
}

type HookConfig struct {
	Url         string `json:"url,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	InsecureSSL string `json:"insecure_ssl,omitempty"`
	Secret      string `json:"secret,omitempty"`
}
