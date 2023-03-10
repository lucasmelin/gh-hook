package cmd

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/cli/go-gh/pkg/config"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func Test_createHook(t *testing.T) {
	tests := []struct {
		name      string
		repo      repository.Repository
		data      Hook
		httpMocks func()
		wantErr   bool
	}{
		{
			name: "success request",
			repo: MockRepo{
				host:  "github.com",
				name:  "test-repo",
				owner: "user1",
			},
			data: Hook{
				Id:     12345678,
				Name:   "web",
				Active: true,
				Events: []string{"push", "pull_request"},
				Config: HookConfig{
					Url:         "https://example.com/webhook",
					ContentType: "json",
					InsecureSSL: "0",
				},
			},
			httpMocks: func() {
				gock.New("https://api.github.com").
					Post("repos/user1/test-repo/hooks").
					BodyString(`{
  "id":12345678,
  "name":"web",
  "active":true,
  "events": [
    "push",
    "pull_request"
  ],
  "config": {
    "url":"https://example.com/webhook",
    "content_type":"json",
    "insecure_ssl":"0"
  }
}`).
					Reply(201).
					JSON(`{
  "type": "Repository",
  "id": 12345678,
  "name": "web",
  "active": true,
  "events": [
    "push",
    "pull_request"
  ],
  "config": {
    "content_type": "json",
    "insecure_ssl": "0",
    "url": "https://example.com/webhook"
  },
  "updated_at": "2019-06-03T00:57:16Z",
  "created_at": "2019-06-03T00:57:16Z",
  "url": "https://api.github.com/repos/user1/test-repo/hooks/12345678",
  "test_url": "https://api.github.com/repos/user1/test-repo/hooks/12345678/test",
  "ping_url": "https://api.github.com/repos/user1/test-repo/hooks/12345678/pings",
  "deliveries_url": "https://api.github.com/repos/user1/test-repo/hooks/12345678/deliveries",
  "last_response": {
    "code": null,
    "status": "unused",
    "message": null
  }
}`)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubConfig(t, testConfig())
			t.Cleanup(gock.Off)
			if tt.httpMocks != nil {
				tt.httpMocks()
			}
			err := createHook(tt.repo, tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("createHook(%v, %v) error = %+v, wantErr %+v", tt.repo, tt.data, err, tt.wantErr)
			}
			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
		})
	}
}

func Test_getEvents(t *testing.T) {
	tests := []struct {
		name      string
		refresh   bool
		httpMocks func()
		want      []string
		wantErr   bool
	}{
		{
			name:    "no refresh, only known events",
			refresh: false,
			want:    knownEvents,
		},
		{
			name:    "successfully refresh list of events",
			refresh: true,
			httpMocks: func() {
				gock.New("https://octokit.github.io").
					Get("webhooks/payload-examples/api.github.com/index.json").
					Reply(200).
					JSON(`[
  {
    "name": "branch_protection_rule",
    "description": "Activity related to a branch protection rule.",
    "actions": ["created", "deleted", "edited"],
    "properties": {
      "rule": {
        "type": "object",
        "description": "The branch protection rule"
      },
      "changes": {
        "type": "object",
        "description": "If the action was edited, the changes to the rule."
      },
      "repository": {
        "type": "object",
        "description": "The [repository](https://docs.github.com/en/rest/reference/repos#get-a-repository) where the event occurred."
      },
      "organization": {
        "type": "object",
        "description": "Webhook payloads contain the [organization](https://docs.github.com/en/rest/reference/orgs#get-an-organization) object when the webhook is configured for an organization or the event occurs from activity in a repository owned by an organization."
      },
      "sender": {
        "type": "object",
        "description": "The user that triggered the event."
      }
    }
}
]`)
			},
			want: []string{"branch_protection_rule"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubConfig(t, testConfig())
			t.Cleanup(gock.Off)
			if tt.httpMocks != nil {
				tt.httpMocks()
			}
			got, err := getEvents(tt.refresh)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getEvents() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEvents() got = %+v, want %+v", got, tt.want)
			}
			assert.Equalf(t, tt.want, got, "getEvents(%v)", tt.refresh)
		})
	}
}

func stubConfig(t *testing.T, cfgStr string) {
	t.Helper()
	old := config.Read
	config.Read = func() (*config.Config, error) {
		return config.ReadFromString(cfgStr), nil
	}
	t.Cleanup(func() {
		config.Read = old
	})
}

func testConfig() string {
	return `
hosts:
  github.com:
    user: user1
    oauth_token: abc123
`
}

func Test_hookFromInput(t *testing.T) {
	tests := []struct {
		name    string
		data    io.Reader
		want    Hook
		wantErr bool
	}{
		{
			name: "basic",
			data: strings.NewReader(`{
  "active": true,
  "events": [
    "push",
    "pull_request"
  ],
  "config": {
    "url": "https://example.com",
    "content_type": "json",
    "insecure_ssl": "0",
    "secret": "somesecretpassphrase"
  }
}`),
			want: Hook{
				Id:     0,
				Name:   "",
				Active: true,
				Events: []string{"push", "pull_request"},
				Config: HookConfig{
					Url:         "https://example.com",
					ContentType: "json",
					InsecureSSL: "0",
					Secret:      "somesecretpassphrase",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hookFromInput(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getEvents() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equalf(t, tt.want, got, "hookFromInput(%v)", tt.data)
		})
	}
}
