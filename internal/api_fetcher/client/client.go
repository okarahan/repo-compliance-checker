package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/okarahan/repo-compliance-checker/internal/common"
)

// Client is the GitHub-specific HTTP client. It wraps the reusable common.Client
// and adds GitHub authentication + content negotiation as default headers.
type Client struct {
	http *common.Client
}

type Options struct {
	// BaseURL defaults to https://api.github.com/.
	BaseURL string

	// HTTPClient defaults to &http.Client{Timeout: 30s}.
	HTTPClient *http.Client
}

func New(token string, opts Options) (*Client, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("github token is empty")
	}

	base := strings.TrimSpace(opts.BaseURL)
	if base == "" {
		base = "https://api.github.com/"
	}

	hc, err := common.New(common.Options{
		BaseURL:    base,
		HTTPClient: opts.HTTPClient,
		Headers: map[string]string{
			"Authorization": "Bearer " + token,
			"Accept":        "application/vnd.github+json",
			"User-Agent":    "repo-compliance-checker",
		},
	})
	if err != nil {
		return nil, err
	}

	return &Client{http: hc}, nil
}

// NewRequest creates an authenticated GitHub API request (GET-style, no body).
// path is appended to the BaseURL and may start with "/".
func (c *Client) NewRequest(method, path string) (*http.Request, error) {
	if c == nil || c.http == nil {
		return nil, fmt.Errorf("client is not initialized")
	}
	return c.http.NewRequest(method, path, nil)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c == nil || c.http == nil {
		return nil, fmt.Errorf("client is not initialized")
	}
	return c.http.Do(req)
}
