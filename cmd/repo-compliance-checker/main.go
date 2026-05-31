package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/okarahan/repo-compliance-checker/internal/analyzer"
	"github.com/okarahan/repo-compliance-checker/internal/analyzer/claude"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
	"github.com/okarahan/repo-compliance-checker/internal/config"
	"github.com/okarahan/repo-compliance-checker/internal/model"
	"github.com/okarahan/repo-compliance-checker/internal/render"
	"github.com/okarahan/repo-compliance-checker/internal/report"
)

// flags holds the parsed command-line configuration.
type flags struct {
	reposPath       string
	allowedPath     string
	manifestMapPath string
	envPath         string
	workdir         string
	outDir          string
	debug           bool
}

// setupFlags defines and parses the command-line flags.
func setupFlags() flags {
	var f flags
	flag.StringVar(&f.reposPath, "repos", "config/repos.json", "path to repos config JSON")
	flag.StringVar(&f.allowedPath, "allowed", "config/allowed_technologies.json", "path to allowed technologies config JSON")
	flag.StringVar(&f.manifestMapPath, "manifest-map", "config/manifest_map.json", "path to manifest map config JSON")
	flag.StringVar(&f.envPath, "env", ".env", "path to .env file (tokens are read from the environment first)")
	flag.StringVar(&f.workdir, "workdir", "", "directory to download manifest files into (default: a temp dir per repo)")
	flag.StringVar(&f.outDir, "out", "reports", "directory to write JSON compliance reports into")
	flag.BoolVar(&f.debug, "debug", false, "enable debug logging on stderr")
	flag.Parse()
	return f
}

func main() {
	f := setupFlags()

	setupLogger(f.debug)

	if err := run(f); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// loadedConfig bundles the three config files the tool needs.
type loadedConfig struct {
	repos       model.ReposConfig
	allowed     model.AllowedTechnologies
	manifestMap model.ManifestMapConfig
}

// loadConfig loads the repos, allowed-technologies and manifest-map config files.
func loadConfig(f flags) (loadedConfig, error) {
	repos, err := config.LoadRepos(f.reposPath)
	if err != nil {
		return loadedConfig{}, err
	}
	allowed, err := config.LoadAllowedTechnologies(f.allowedPath)
	if err != nil {
		return loadedConfig{}, err
	}
	manifestMap, err := config.LoadManifestMap(f.manifestMapPath)
	if err != nil {
		return loadedConfig{}, err
	}

	slog.Debug("loaded repos config", "path", f.reposPath, "count", len(repos.Repos), "repos", repos.Repos)
	slog.Debug("loaded allowed technologies config", "path", f.allowedPath, "allowed", allowed)
	slog.Debug("loaded manifest map config", "path", f.manifestMapPath, "manifest_map", manifestMap)

	return loadedConfig{repos: repos, allowed: allowed, manifestMap: manifestMap}, nil
}

// createGithubClient reads the GitHub token (env first, then envPath) and builds
// the GitHub API client.
func createGithubClient(envPath string) (*client.Client, error) {
	token, err := client.ReadGitHubToken(envPath)
	if err != nil {
		return nil, fmt.Errorf("github token: %w", err)
	}
	return client.New(token, client.Options{})
}

// createClaudeClient reads the Anthropic API key (env first, then envPath) and
// builds the Claude classifier client.
func createClaudeClient(envPath string) (*claude.Client, error) {
	apiKey, err := claude.ReadAPIKey(envPath)
	if err != nil {
		return nil, fmt.Errorf("anthropic api key: %w", err)
	}
	return claude.New(apiKey, claude.Options{})
}

// setupLogger configures the default slog logger to write to stderr, keeping stdout
// reserved for the JSON result output.
func setupLogger(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}

func run(f flags) error {
	cfg, err := loadConfig(f)
	if err != nil {
		return err
	}
	repos, allowed, manifestMap := cfg.repos, cfg.allowed, cfg.manifestMap

	gh, err := createGithubClient(f.envPath)
	if err != nil {
		return err
	}
	classifier, err := createClaudeClient(f.envPath)
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, repo := range repos.Repos {
		result, languages, err := analyzeRepo(ctx, gh, classifier, repo.Slug, repo.Ref, manifestMap, allowed, f.workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %q: %v\n", repo.Slug, err)
			continue
		}

		if err := generateReport(repo.Slug, result, languages, allowed, f.outDir); err != nil {
			return err
		}
	}

	return nil
}

// generateReport builds the compliance report for a repo, writes the JSON and HTML
// outputs into outDir, and logs/prints a short summary.
func generateReport(
	slug string,
	result model.AnalysisResult,
	languages map[string]int64,
	allowed model.AllowedTechnologies,
	outDir string,
) error {
	rep := report.Build(slug, result, languages, allowed)

	path, err := report.Write(outDir, rep)
	if err != nil {
		return err
	}
	htmlPath, err := render.Write(outDir, rep)
	if err != nil {
		return err
	}

	c := rep.Conclusion
	slog.Info("report written",
		"repo", slug, "path", path, "html", htmlPath,
		"detected", c.DetectedCount, "allowed", c.AllowedCount,
		"language_pct", c.Categories.Language.CompliancePercentage,
		"framework_pct", c.Categories.Framework.CompliancePercentage,
		"utility_pct", c.Categories.Utility.CompliancePercentage,
		"overall_pct", c.OverallCompliancePercentage, "compliant", c.Compliant,
	)
	fmt.Printf("%s: overall %.1f%% (lang %.1f%% / framework %.1f%% / util %.1f%%), compliant=%t -> %s | %s\n",
		slug, c.OverallCompliancePercentage,
		c.Categories.Language.CompliancePercentage,
		c.Categories.Framework.CompliancePercentage,
		c.Categories.Utility.CompliancePercentage,
		c.Compliant, path, htmlPath)

	return nil
}

func analyzeRepo(
	ctx context.Context,
	gh *client.Client,
	classifier *claude.Client,
	slug, ref string,
	manifestMap model.ManifestMapConfig,
	allowed model.AllowedTechnologies,
	workdir string,
) (model.AnalysisResult, map[string]int64, error) {
	owner, name, err := splitSlug(slug)
	if err != nil {
		return model.AnalysisResult{}, nil, err
	}

	repoFetchResult, err := api_fetcher.FetchRepo(ctx, gh, owner, name, ref, manifestMap, workdir)
	if err != nil {
		return model.AnalysisResult{}, nil, fmt.Errorf("fetch: %w", err)
	}
	slog.Debug("fetched repo",
		"owner", owner, "repo", name,
		"languages", repoFetchResult.Metadata.Languages,
		"topics", repoFetchResult.Metadata.Topics,
		"downloaded", repoFetchResult.Manifest.Downloaded,
		"missing", repoFetchResult.Manifest.Missing,
		"dir", repoFetchResult.Manifest.Dir,
	)

	deps, err := analyzer.DetectDependencies(repoFetchResult, manifestMap)
	if err != nil {
		return model.AnalysisResult{}, nil, fmt.Errorf("detect dependencies: %w", err)
	}
	slog.Debug("detected dependencies", "owner", owner, "repo", name, "count", len(deps), "dependencies", deps)

	result, err := classifier.Classify(ctx, deps, allowed)
	if err != nil {
		return model.AnalysisResult{}, nil, fmt.Errorf("classify: %w", err)
	}
	return result, repoFetchResult.Metadata.Languages, nil
}

func splitSlug(slug string) (owner, repo string, err error) {
	owner, repo, ok := strings.Cut(strings.TrimSpace(slug), "/")
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	if !ok || owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid slug, expected owner/repo")
	}
	return owner, repo, nil
}
