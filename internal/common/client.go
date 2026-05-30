package common

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a small, reusable HTTP client for talking to JSON REST APIs.
//
// It is intentionally generic: authentication and content negotiation are expressed
// as default headers, so the same client works for GitHub, the Anthropic API, etc.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	headers    map[string]string
}

type Options struct {
	// BaseURL is the absolute API base URL (required), e.g. "https://api.github.com/".
	BaseURL string

	// HTTPClient defaults to &http.Client{Timeout: 30s}.
	HTTPClient *http.Client

	// Headers are applied to every request created via NewRequest
	// (e.g. Authorization, Accept, anthropic-version).
	Headers map[string]string
}

func New(opts Options) (*Client, error) {
	base := strings.TrimSpace(opts.BaseURL)
	if base == "" {
		return nil, fmt.Errorf("base url is required")
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

	headers := make(map[string]string, len(opts.Headers))
	for k, v := range opts.Headers {
		headers[k] = v
	}

	return &Client{
		baseURL:    u,
		httpClient: hc,
		headers:    headers,
	}, nil
}

// NewRequest creates a request against the base URL with the configured default headers.
// path is appended to the BaseURL and may start with "/". body may be nil.
func (c *Client) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	if c == nil || c.httpClient == nil || c.baseURL == nil {
		return nil, fmt.Errorf("client is not initialized")
	}

	u := c.baseURL.ResolveReference(&url.URL{Path: strings.TrimSpace(path)})
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}
