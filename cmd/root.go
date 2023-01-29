package cmd

import (
	"fmt"
	"os"

	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/repository"
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

func getRepo(cmd *cobra.Command) (repository.Repository, error) {
	var repo repository.Repository
	var err error
	repoOverride, err := cmd.Flags().GetString("repo")
	if err != nil {
		return nil, fmt.Errorf("could not parse repo flag: %w\n", err)
	}
	if repoOverride != "" {
		repo, err = repository.Parse(repoOverride)
	} else {
		repo, err = gh.CurrentRepository()
	}
	if err != nil {
		return nil, fmt.Errorf("could not determine the repo to use: %w\n", err)
	}
	return repo, nil
}

type Event struct {
	Name string `json:"name"`
}

type Hook struct {
	Id     int        `json:"id,omitempty"`
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

// Some events are not available for repositories.
// See: https://docs.github.com/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#installation_repositories
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
	// "github_app_authorization", Not available for repositories.
	"gollum",
	// "installation", Not available for repositories.
	// "installation_repositories", Not available for repositories.
	"issue_comment",
	"issues",
	"label",
	// "marketplace_purchase", Not available for repositories.
	"member",
	// "membership", Not available for repositories.
	// "merge_group", Not available for repositories.
	"meta",
	"milestone",
	// "organization", Not available for repositories.
	// "org_block", Not available for repositories.
	"package",
	"page_build",
	"ping",
	"project",
	"project_card",
	"project_column",
	// "projects_v2_item", Not available for repositories.
	"public",
	"pull_request",
	"pull_request_review",
	"pull_request_review_comment",
	"pull_request_review_thread",
	"push",
	"release",
	// "repository_dispatch", Not available for repositories.
	"repository",
	"repository_import",
	"repository_vulnerability_alert",
	// "security_advisory", Not available for repositories.
	// "sponsorship", Not available for repositories.
	"star",
	"status",
	// "team", Not available for repositories.
	"team_add",
	"watch",
	// "workflow_dispatch", Not available for repositories.
	"workflow_job",
	"workflow_run",
}
