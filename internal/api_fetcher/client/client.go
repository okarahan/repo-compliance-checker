package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	token      string
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
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("base url must be absolute, got %q", base)
	}

	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		baseURL:    u,
		httpClient: hc,
		token:      token,
	}, nil
}

// NewRequest creates an authenticated GitHub API request.
// path is appended to the BaseURL and may start with "/".
func (c *Client) NewRequest(method, path string) (*http.Request, error) {
	if c == nil || c.httpClient == nil || c.baseURL == nil {
		return nil, fmt.Errorf("client is not initialized")
	}

	u := c.baseURL.ResolveReference(&url.URL{Path: strings.TrimSpace(path)})
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "repo-compliance-checker")
	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

