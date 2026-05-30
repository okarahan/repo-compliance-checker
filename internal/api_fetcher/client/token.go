package client

import (
	"github.com/okarahan/repo-compliance-checker/internal/common"
)

const (
	// EnvGitHubToken is the env var name used by GitHub authentication.
	EnvGitHubToken = "GITHUB_TOKEN"
)

// ReadGitHubToken reads GITHUB_TOKEN from the process environment, falling back to
// the provided .env file path. The process environment takes precedence.
//
// This function intentionally does not validate token format; it only loads it.
func ReadGitHubToken(envFilePath string) (string, error) {
	return common.ReadEnvValue(EnvGitHubToken, envFilePath)
}
