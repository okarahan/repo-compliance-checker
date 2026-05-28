package client_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
)

func TestReadGitHubToken_fromEnvFile(t *testing.T) {
	t.Setenv(client.EnvGitHubToken, "")

	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("GITHUB_TOKEN=abc123\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := client.ReadGitHubToken(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != "abc123" {
		t.Fatalf("token=%q, want %q", got, "abc123")
	}
}

func TestReadGitHubToken_envTakesPrecedence(t *testing.T) {
	t.Setenv(client.EnvGitHubToken, "from-env")

	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("GITHUB_TOKEN=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := client.ReadGitHubToken(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != "from-env" {
		t.Fatalf("token=%q, want %q", got, "from-env")
	}
}

func TestReadGitHubToken_missing(t *testing.T) {
	t.Setenv(client.EnvGitHubToken, "")

	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("# empty\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := client.ReadGitHubToken(envPath)
	if err == nil {
		t.Fatal("expected error")
	}
}

