package api_fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type repoResponse struct {
	Topics []string `json:"topics"`
}

// TODO: common response handling and common request generation, GetTopics should have only
// the API call and the response mapping
func (c *Client) GetTopics(ctx context.Context, owner, repo string) ([]string, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	path := fmt.Sprintf("/repos/%s/%s", url.PathEscape(owner), url.PathEscape(repo))
	req, err := c.NewRequest(http.MethodGet, path)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github repo api: status=%s body=%s", resp.Status, strings.TrimSpace(string(b)))
	}

	var out repoResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode repo response: %w", err)
	}
	if out.Topics == nil {
		return []string{}, nil
	}
	return out.Topics, nil
}
