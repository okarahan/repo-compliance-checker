package analyzer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/analyzer"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/endpoints"
	"github.com/okarahan/repo-compliance-checker/internal/model"
)

func TestDetectDependencies(t *testing.T) {
	dir := t.TempDir()

	goMod := "module example\n\ngo 1.22\n\nrequire github.com/labstack/echo/v4 v4.11.0\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}

	reqs := "# this is a comment\ntestcontainers==4.0\nruff==0.4.0\n"
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(reqs), 0o644); err != nil {
		t.Fatal(err)
	}

	result := api_fetcher.RepoFetchResult{
		Metadata: api_fetcher.RepoMetadata{
			Languages: endpoints.LanguagesResponse{"Go": 100, "Python": 50},
		},
		Manifest: api_fetcher.ManifestDownload{
			Dir:        dir,
			Downloaded: []string{"go.mod", "requirements.txt"},
		},
	}

	manifestCfg := model.ManifestMapConfig{
		ByLanguage: map[string]model.ManifestTargets{
			"go":     {Files: []string{"go.mod", "go.sum"}},
			"python": {Files: []string{"requirements.txt"}},
		},
	}

	findings, err := analyzer.DetectDependencies(result, manifestCfg)
	if err != nil {
		t.Fatal(err)
	}

	var gotGo, gotTestcontainers, gotRuff, gotComment bool
	for _, f := range findings {
		switch {
		case f.File == "go.mod" && f.Name == "github.com/labstack/echo/v4":
			gotGo = true
		case f.File == "requirements.txt" && f.Name == "testcontainers":
			gotTestcontainers = true
		case f.File == "requirements.txt" && f.Name == "ruff":
			gotRuff = true
		case f.Line == "# this is a comment":
			gotComment = true
		}
	}

	if !gotGo {
		t.Errorf("expected go.mod require line, findings=%v", findings)
	}
	if !gotTestcontainers || !gotRuff {
		t.Errorf("expected python deps, findings=%v", findings)
	}
	if gotComment {
		t.Errorf("comment line should be skipped, findings=%v", findings)
	}
}
