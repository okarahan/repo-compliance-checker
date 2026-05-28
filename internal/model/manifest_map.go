package model

// ManifestMapConfig defines which files/directories should be fetched (repo-root relative)
// to detect technologies in a repository.
//
// This is designed to be usable both as a local JSON config and as a future REST payload.
type ManifestMapConfig struct {
	Version    int                       `json:"version"`
	ByLanguage map[string]ManifestTargets `json:"by_language"`
	Global     ManifestTargets            `json:"global"`
}

type ManifestTargets struct {
	Files       []string `json:"files"`
	Directories []string `json:"directories"`
}

