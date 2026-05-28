package api_fetcher

import (
	"context"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/endpoints"
)

// RepoMetadata is the minimal API-derived information we use for MVP decisions.
type RepoMetadata struct {
	Languages endpoints.LanguagesResponse
	Topics    []string
}

// FetchRepoMetadata creates a GitHub API client (token already loaded) and fetches
// Linguist languages + repo topics for a given owner/repo.
//
// Naming rationale: this is metadata-like information (not full file contents).
func FetchRepoMetadata(ctx context.Context, c *client.Client, owner, repo string) (RepoMetadata, error) {
	langs, err := endpoints.GetLanguages(ctx, c, owner, repo)
	if err != nil {
		return RepoMetadata{}, err
	}
	topics, err := endpoints.GetTopics(ctx, c, owner, repo)
	if err != nil {
		return RepoMetadata{}, err
	}
	return RepoMetadata{Languages: langs, Topics: topics}, nil
}

