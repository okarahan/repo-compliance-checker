package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReadEnvValue returns the value for key, preferring the process environment and
// falling back to the given .env file. It returns an error if the key is not found
// in either place. It intentionally does not validate the value's format.
func ReadEnvValue(key, envFilePath string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("env key is empty")
	}

	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v, nil
	}

	if strings.TrimSpace(envFilePath) == "" {
		return "", fmt.Errorf("env file path is empty and %s is not set", key)
	}

	m, err := ReadDotEnv(envFilePath)
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(m[key])
	if v == "" {
		return "", fmt.Errorf("%s not found in %s", key, envFilePath)
	}
	return v, nil
}

// ReadDotEnv parses a simple .env file (KEY=VALUE per line, # comments) into a map.
func ReadDotEnv(path string) (map[string]string, error) {
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
