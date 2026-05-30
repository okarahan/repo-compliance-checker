package claude

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/okarahan/repo-compliance-checker/internal/common"
)

const (
	// DefaultBaseURL is the Anthropic API base URL.
	DefaultBaseURL = "https://api.anthropic.com/"

	// anthropicVersion is the required API version header value.
	anthropicVersion = "2023-06-01"

	// DefaultModel is used when Options.Model is empty. Adjust to a currently
	// available model slug if needed.
	DefaultModel = "claude-3-5-sonnet-latest"

	// defaultMaxTokens caps the size of the model's response.
	defaultMaxTokens = 1024
)

// Client is a minimal Anthropic Messages API client built on common.Client.
type Client struct {
	http      *common.Client
	model     string
	maxTokens int
}

type Options struct {
	// BaseURL defaults to DefaultBaseURL.
	BaseURL string
	// Model defaults to DefaultModel.
	Model string
	// MaxTokens defaults to defaultMaxTokens.
	MaxTokens int
	// HTTPClient defaults to &http.Client{Timeout: 30s}.
	HTTPClient *http.Client
}

func New(apiKey string, opts Options) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic api key is empty")
	}

	base := strings.TrimSpace(opts.BaseURL)
	if base == "" {
		base = DefaultBaseURL
	}

	hc, err := common.New(common.Options{
		BaseURL:    base,
		HTTPClient: opts.HTTPClient,
		Headers: map[string]string{
			"x-api-key":         apiKey,
			"anthropic-version": anthropicVersion,
			"content-type":      "application/json",
		},
	})
	if err != nil {
		return nil, err
	}

	model := strings.TrimSpace(opts.Model)
	if model == "" {
		model = DefaultModel
	}
	maxTokens := opts.MaxTokens
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}

	return &Client{http: hc, model: model, maxTokens: maxTokens}, nil
}
