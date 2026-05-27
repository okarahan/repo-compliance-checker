package model

// ReposConfig is the request/response body for listing repositories to analyze.
// Same shape for local JSON files and future REST API payloads.
type ReposConfig struct {
	Repos []RepoTarget `json:"repos"`
}

// RepoTarget identifies a GitHub repository to analyze.
type RepoTarget struct {
	// Slug is the GitHub repository in org/repo form (e.g. "golang/go").
	Slug string `json:"slug"`

	// Ref is an optional branch, tag, or commit (default interpreted by fetcher, e.g. "main").
	Ref string `json:"ref,omitempty"`
}
