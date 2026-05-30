package claude

import (
	"github.com/okarahan/repo-compliance-checker/internal/common"
)

// EnvAPIKey is the env var name used for Anthropic authentication.
const EnvAPIKey = "ANTHROPIC_API_KEY"

// ReadAPIKey reads ANTHROPIC_API_KEY from the process environment, falling back to
// the provided .env file path. The process environment takes precedence.
func ReadAPIKey(envFilePath string) (string, error) {
	return common.ReadEnvValue(EnvAPIKey, envFilePath)
}
