package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	// EnvGitHubToken is the env var name used by GitHub authentication.
	EnvGitHubToken = "GITHUB_TOKEN"
)

// ReadGitHubToken reads GITHUB_TOKEN from the provided .env file path.
// If the token is present in the process environment, it takes precedence.
//
// This function intentionally does not validate token format; it only loads it.
func ReadGitHubToken(envFilePath string) (string, error) {
	if v := strings.TrimSpace(os.Getenv(EnvGitHubToken)); v != "" {
		return v, nil
	}
	if strings.TrimSpace(envFilePath) == "" {
		return "", fmt.Errorf("env file path is empty and %s is not set", EnvGitHubToken)
	}
	m, err := readDotEnv(envFilePath)
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(m[EnvGitHubToken])
	if v == "" {
		return "", fmt.Errorf("%s not found in %s", EnvGitHubToken, envFilePath)
	}
	return v, nil
}

func readDotEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open .env file %q: %w", path, err)
	}
	defer f.Close()

	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" {
			continue
		}
		out[k] = v
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan .env file %q: %w", path, err)
	}
	return out, nil
}

