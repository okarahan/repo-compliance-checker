package api_fetcher_test

import (
	"context"
	"os"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher"
)

func TestIntegration_GetTopics(t *testing.T) {
	if os.Getenv("RCC_INTEGRATION") != "1" {
		t.Skip("set RCC_INTEGRATION=1 to run integration tests")
	}

	token := os.Getenv(api_fetcher.EnvGitHubToken)
	if token == "" {
		t.Skipf("set %s to run integration tests", api_fetcher.EnvGitHubToken)
	}

	c, err := api_fetcher.NewClient(token, api_fetcher.ClientOptions{})
	if err != nil {
		t.Fatal(err)
	}

	topics, err := c.GetTopics(context.Background(), "golang", "go")
	if err != nil {
		t.Fatal(err)
	}
	// Topics may be empty depending on repo settings; just ensure call succeeds.
	_ = topics
}

