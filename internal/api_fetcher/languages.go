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

// LanguagesResponse maps language name to number of bytes of code.
// This is GitHub Linguist's output for the repo's default branch.
type LanguagesResponse map[string]int64

func (c *Client) GetLanguages(ctx context.Context, owner, repo string) (LanguagesResponse, error) {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	path := fmt.Sprintf("/repos/%s/%s/languages", url.PathEscape(owner), url.PathEscape(repo))
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
		return nil, fmt.Errorf("github languages api: status=%s body=%s", resp.Status, strings.TrimSpace(string(b)))
	}

	var out LanguagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode languages response: %w", err)
	}
	return out, nil
}

