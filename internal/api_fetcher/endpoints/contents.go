package endpoints

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
)

// ErrFileNotFound is returned by GetFileContent when the path does not exist
// at the given ref (HTTP 404). Callers can treat this as a "gap", not a fatal error.
var ErrFileNotFound = errors.New("file not found")

type contentsResponse struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
}

// GetFileContent fetches a single file via the GitHub Contents API and returns its
// decoded bytes. ref may be empty (GitHub then uses the default branch).
//
// Returns ErrFileNotFound if the file does not exist at the given path/ref.
func GetFileContent(ctx context.Context, c *client.Client, owner, repo, ref, path string) ([]byte, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	path = strings.TrimSpace(path)
	if owner == "" || repo == "" || path == "" {
		return nil, fmt.Errorf("owner, repo and path are required")
	}
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}

	apiPath := fmt.Sprintf("/repos/%s/%s/contents/%s", url.PathEscape(owner), url.PathEscape(repo), encodePath(path))

	req, err := c.NewRequest(http.MethodGet, apiPath)
	if err != nil {
		return nil, err
	}
	if ref := strings.TrimSpace(ref); ref != "" {
		req.URL.RawQuery = url.Values{"ref": {ref}}.Encode()
	}
	req = req.WithContext(ctx)

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%s: %w", path, ErrFileNotFound)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github contents api: status=%s body=%s", resp.Status, strings.TrimSpace(string(b)))
	}

	var out contentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode contents response: %w", err)
	}

	if out.Type != "file" {
		return nil, fmt.Errorf("path %q is not a file (type=%q)", path, out.Type)
	}

	switch out.Encoding {
	case "base64":
		// GitHub may wrap base64 content with newlines.
		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(out.Content, "\n", ""))
		if err != nil {
			return nil, fmt.Errorf("decode base64 content for %q: %w", path, err)
		}
		return decoded, nil
	case "", "none":
		return nil, fmt.Errorf("file %q has unsupported/empty encoding %q (too large for contents api?)", path, out.Encoding)
	default:
		return nil, fmt.Errorf("file %q has unexpected encoding %q", path, out.Encoding)
	}
}

// encodePath escapes each path segment but keeps the slashes intact.
func encodePath(p string) string {
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		parts[i] = url.PathEscape(seg)
	}
	return strings.Join(parts, "/")
}
