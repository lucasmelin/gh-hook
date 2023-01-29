package cmd

import (
	"testing"

	"github.com/cli/go-gh/pkg/repository"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func Test_deleteHooks(t *testing.T) {
	tests := []struct {
		name      string
		repo      repository.Repository
		deleteIds []string
		httpMocks func()
		wantErr   bool
	}{
		{
			name: "success request single hook",
			repo: MockRepo{
				host:  "github.com",
				name:  "test-repo",
				owner: "lucasmelin",
			},
			httpMocks: func() {
				gock.New("https://api.github.com").
					Delete("repos/lucasmelin/test-repo/hooks/12365678").
					Reply(204)
			},
			deleteIds: []string{"12365678"},
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
			err := deleteHooks(tt.repo, tt.deleteIds)
			if (err != nil) != tt.wantErr {
				t.Fatalf("deleteHooks(%v, %v) error = %v, wantErr %v", tt.repo, tt.deleteIds, err, tt.wantErr)
			}
			assert.True(t, gock.IsDone(), printPendingMocks(gock.Pending()))
		})
	}
}
