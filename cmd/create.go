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

type Event struct {
	Name string `json:"name"`
}

type Hook struct {
	Id     int      `json:"id,omitempty"`
	Name   string   `json:"name,omitempty"`
	Active bool     `json:"active,omitempty"`
	Events []string `json:"events,omitempty"`
	Config HookConfig
}

type HookConfig struct {
	Url         string `json:"url,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	InsecureSSL string `json:"insecure_ssl,omitempty"`
	Secret      string `json:"secret,omitempty"`
}

var knownEvents = []string{
	"branch_protection_rule",
	"check_run",
	"check_suite",
	"code_scanning_alert",
	"commit_comment",
	"create",
	"delete",
	"dependabot_alert",
	"deploy_key",
	"deployment",
	"deployment_status",
	"discussion",
	"discussion_comment",
	"fork",
	"github_app_authorization",
	"gollum",
	"installation",
	"installation_repositories",
	"issue_comment",
	"issues",
	"label",
	"marketplace_purchase",
	"member",
	"membership",
	"merge_group",
	"meta",
	"milestone",
	"organization",
	"org_block",
	"package",
	"page_build",
	"ping",
	"project",
	"project_card",
	"project_column",
	"projects_v2_item",
	"public",
	"pull_request",
	"pull_request_review",
	"pull_request_review_comment",
	"pull_request_review_thread",
	"push",
	"release",
	"repository_dispatch",
	"repository",
	"repository_import",
	"repository_vulnerability_alert",
	"security_advisory",
	"sponsorship",
	"star",
	"status",
	"team",
	"team_add",
	"watch",
	"workflow_dispatch",
	"workflow_job",
	"workflow_run",
}

var repoOverride string
var refreshEvents bool

func init() {
	createCmd.Flags().StringVarP(&repoOverride, "repo", "", "", "Specify a repository. If omitted, uses current repository")
	createCmd.Flags().BoolVarP(&refreshEvents, "refresh-events", "", false, "Download the list of events from octokit.github.io/webhooks/ By default, will use a hardcoded list of known events.")
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
		}
		client, err := gh.RESTClient(&hookOpts)
		if err != nil {
			return fmt.Errorf("error creating REST client: %w\n", err)
		}
		value := Hook{
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
			return fmt.Errorf("error converting resoponses to JSON: %w\n", err)
		}
		apiUrl := fmt.Sprintf("repos/%s/%s/hooks", repo.Owner(), repo.Name())
		err = client.Post(apiUrl, bytes.NewBuffer(jsonValue), nil)
		if err != nil {
			return fmt.Errorf("error creating new webhook: %w\n", err)
		}
		fmt.Println("successfully created hook")
		return nil
	},
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
	defer resp.Body.Close()
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
