package endpoints_test

import (
	"context"
	"os"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/endpoints"
)

func TestIntegration_GetTopics(t *testing.T) {
	if os.Getenv("RCC_INTEGRATION") != "1" {
		t.Skip("set RCC_INTEGRATION=1 to run integration tests")
	}

	token := os.Getenv(client.EnvGitHubToken)
	if token == "" {
		t.Skipf("set %s to run integration tests", client.EnvGitHubToken)
	}

	c, err := client.New(token, client.Options{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = endpoints.GetTopics(context.Background(), c, "golang", "go")
	if err != nil {
		t.Fatal(err)
	}
}

