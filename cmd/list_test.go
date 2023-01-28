package cmd

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cli/go-gh/pkg/repository"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

type MockRepo struct {
	host  string
	name  string
	owner string
}

func (mr MockRepo) Host() string {
	return mr.host
}

func (mr MockRepo) Name() string {
	return mr.name
}

func (mr MockRepo) Owner() string {
	return mr.owner
}

func Test_getWebhooks(t *testing.T) {
	tests := []struct {
		name      string
		repo      repository.Repository
		httpMocks func()
		want      []Hook
		wantErr   bool
	}{
		{
			name: "success request empty response",
			repo: MockRepo{
				host:  "github.com",
				name:  "test-repo",
				owner: "lucasmelin",
			},
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("repos/lucasmelin/test-repo/hooks").
					Reply(204)
			},
			want: []Hook{},
		},
		{
			name: "success request single response",
			repo: MockRepo{
				host:  "github.com",
				name:  "Hello-World",
				owner: "octocat",
			},
			httpMocks: func() {
				gock.New("https://api.github.com").
					Get("repos/octocat/Hello-World/hooks").
					Reply(200).
					JSON(`[
  {
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
    "url": "https://api.github.com/repos/octocat/Hello-World/hooks/12345678",
    "test_url": "https://api.github.com/repos/octocat/Hello-World/hooks/12345678/test",
    "ping_url": "https://api.github.com/repos/octocat/Hello-World/hooks/12345678/pings",
    "deliveries_url": "https://api.github.com/repos/octocat/Hello-World/hooks/12345678/deliveries",
    "last_response": {
      "code": null,
      "status": "unused",
      "message": null
    }
  }
]`)
			},
			want: []Hook{
				{
					Id:     12345678,
					Name:   "web",
					Active: true,
					Events: []string{
						"push",
						"pull_request",
					},
					Config: HookConfig{
						ContentType: "json",
						InsecureSSL: "0",
						Url:         "https://example.com/webhook",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "support enterprise hosts",
			repo: MockRepo{
				host:  "enterprise.com",
				name:  "test-repo",
				owner: "lucasmelin",
			},
			httpMocks: func() {
				gock.New("https://enterprise.com").
					Get("repos/lucasmelin/test-repo/hooks").
					Reply(204)
			},
			want: []Hook{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(gock.Off)
			if tt.httpMocks != nil {
				tt.httpMocks()
			}
			if tt.repo.Host() != "github.com" {
				t.Setenv("GH_ENTERPRISE_TOKEN", "mock_token")
			}
			got, err := getWebhooks(tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getWebhooks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getWebhooks() got = %+v, want %+v", got, tt.want)
			}
			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
		})
	}
}

func printPendingMocks(mocks []gock.Mock) string {
	paths := []string{}
	for _, mock := range mocks {
		paths = append(paths, mock.Request().URLStruct.String())
	}
	return fmt.Sprintf("%d unmatched mocks: %s", len(paths), strings.Join(paths, ", "))
}
