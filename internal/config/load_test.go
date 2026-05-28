package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/config"
)

func TestLoadRepos(t *testing.T) {
	path := filepath.Join("testdata", "repos.json")
	cfg, err := config.LoadRepos(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("repos len = %d, want 1", len(cfg.Repos))
	}
	if cfg.Repos[0].Slug != "org/example" {
		t.Fatalf("slug = %q, want org/example", cfg.Repos[0].Slug)
	}
	if cfg.Repos[0].Ref != "main" {
		t.Fatalf("ref = %q, want main", cfg.Repos[0].Ref)
	}
}

func TestLoadAllowedTechnologies(t *testing.T) {
	path := filepath.Join("testdata", "allowed_technologies.json")
	cfg, err := config.LoadAllowedTechnologies(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ProgrammingLanguages) != 1 || cfg.ProgrammingLanguages[0] != "Go" {
		t.Fatalf("programming_languages = %v, want [Go]", cfg.ProgrammingLanguages)
	}
	if len(cfg.Frameworks) != 1 || cfg.Frameworks[0] != "Spring" {
		t.Fatalf("frameworks = %v, want [Spring]", cfg.Frameworks)
	}
	if len(cfg.Utilities) != 1 || cfg.Utilities[0] != "Docker" {
		t.Fatalf("utilities = %v, want [Docker]", cfg.Utilities)
	}
}

func TestLoadRepos_invalidJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "repos-*.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("{ not json"); err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = config.LoadRepos(f.Name())
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestLoadRepos_missingFile(t *testing.T) {
	_, err := config.LoadRepos(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadManifestMap(t *testing.T) {
	path := filepath.Join("..", "..", "config", "manifest_map.json")
	cfg, err := config.LoadManifestMap(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 {
		t.Fatalf("version=%d, want 1", cfg.Version)
	}
	if len(cfg.ByLanguage) == 0 {
		t.Fatal("expected by_language to be non-empty")
	}
	if _, ok := cfg.ByLanguage["go"]; !ok {
		t.Fatal("expected go entry in by_language")
	}
	if len(cfg.Global.Files) == 0 {
		t.Fatal("expected global.files to be non-empty")
	}
}
