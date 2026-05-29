package api_fetcher

import (
	"context"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/endpoints"
	"github.com/okarahan/repo-compliance-checker/internal/model"
)

// RepoMetadata is the minimal API-derived information we use for MVP decisions.
type RepoMetadata struct {
	Languages endpoints.LanguagesResponse
	Topics    []string
}

// RepoFetchResult aggregates everything the analyzer needs for one repo:
// the API-derived metadata plus where the downloaded manifest files live on disk.
type RepoFetchResult struct {
	Metadata RepoMetadata
	Manifest ManifestDownload
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

// FetchRepo orchestrates a full fetch for one repo:
//  1. fetch metadata (languages + topics) via the GitHub API,
//  2. download the language-gated manifest files into destDir,
//  3. return both aggregated in a RepoFetchResult.
//
// mm is the already-loaded manifest map (loading stays in the config package).
// If destDir is empty, DownloadManifestFiles creates a temp directory.
func FetchRepo(ctx context.Context, c *client.Client, owner, repo, ref string, mm model.ManifestMapConfig, destDir string) (RepoFetchResult, error) {
	meta, err := FetchRepoMetadata(ctx, c, owner, repo)
	if err != nil {
		return RepoFetchResult{}, err
	}

	dl, err := DownloadManifestFiles(ctx, c, owner, repo, ref, meta.Languages, mm, destDir)
	if err != nil {
		return RepoFetchResult{}, err
	}

	return RepoFetchResult{Metadata: meta, Manifest: dl}, nil
}

