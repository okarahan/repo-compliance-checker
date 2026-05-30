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
	"github.com/okarahan/repo-compliance-checker/internal/report"
)

func main() {
	reposPath := flag.String("repos", "config/repos.json", "path to repos config JSON")
	allowedPath := flag.String("allowed", "config/allowed_technologies.json", "path to allowed technologies config JSON")
	manifestMapPath := flag.String("manifest-map", "config/manifest_map.json", "path to manifest map config JSON")
	envPath := flag.String("env", ".env", "path to .env file (tokens are read from the environment first)")
	workdir := flag.String("workdir", "", "directory to download manifest files into (default: a temp dir per repo)")
	outDir := flag.String("out", "reports", "directory to write JSON compliance reports into")
	debug := flag.Bool("debug", true, "enable debug logging on stderr")
	flag.Parse()

	setupLogger(*debug)

	if err := run(*reposPath, *allowedPath, *manifestMapPath, *envPath, *workdir, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
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

func run(reposPath, allowedPath, manifestMapPath, envPath, workdir, outDir string) error {
	repos, err := config.LoadRepos(reposPath)
	if err != nil {
		return err
	}
	allowed, err := config.LoadAllowedTechnologies(allowedPath)
	if err != nil {
		return err
	}
	manifestMap, err := config.LoadManifestMap(manifestMapPath)
	if err != nil {
		return err
	}

	slog.Debug("loaded repos config", "path", reposPath, "count", len(repos.Repos), "repos", repos.Repos)
	slog.Debug("loaded allowed technologies config", "path", allowedPath, "allowed", allowed)
	slog.Debug("loaded manifest map config", "path", manifestMapPath, "manifest_map", manifestMap)

	githubToken, err := client.ReadGitHubToken(envPath)
	if err != nil {
		return fmt.Errorf("github token: %w", err)
	}
	apiKey, err := claude.ReadAPIKey(envPath)
	if err != nil {
		return fmt.Errorf("anthropic api key: %w", err)
	}

	gh, err := client.New(githubToken, client.Options{})
	if err != nil {
		return err
	}
	classifier, err := claude.New(apiKey, claude.Options{})
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, repo := range repos.Repos {
		owner, name, err := splitSlug(repo.Slug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %q: %v\n", repo.Slug, err)
			continue
		}

		result, languages, err := analyzeRepo(ctx, gh, classifier, owner, name, repo.Ref, manifestMap, allowed, workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %q: %v\n", repo.Slug, err)
			continue
		}

		rep := report.Build(repo.Slug, result, languages, allowed)
		path, err := report.Write(outDir, rep)
		if err != nil {
			return err
		}

		c := rep.Conclusion
		slog.Info("report written",
			"repo", repo.Slug, "path", path,
			"detected", c.DetectedCount, "allowed", c.AllowedCount,
			"language_pct", c.Categories.Language.CompliancePercentage,
			"framework_pct", c.Categories.Framework.CompliancePercentage,
			"utility_pct", c.Categories.Utility.CompliancePercentage,
			"overall_pct", c.OverallCompliancePercentage, "compliant", c.Compliant,
		)
		fmt.Printf("%s: overall %.1f%% (lang %.1f%% / framework %.1f%% / util %.1f%%), compliant=%t -> %s\n",
			repo.Slug, c.OverallCompliancePercentage,
			c.Categories.Language.CompliancePercentage,
			c.Categories.Framework.CompliancePercentage,
			c.Categories.Utility.CompliancePercentage,
			c.Compliant, path)
	}

	return nil
}

func analyzeRepo(
	ctx context.Context,
	gh *client.Client,
	classifier *claude.Client,
	owner, name, ref string,
	manifestMap model.ManifestMapConfig,
	allowed model.AllowedTechnologies,
	workdir string,
) (model.AnalysisResult, map[string]int64, error) {
	fetched, err := api_fetcher.FetchRepo(ctx, gh, owner, name, ref, manifestMap, workdir)
	if err != nil {
		return model.AnalysisResult{}, nil, fmt.Errorf("fetch: %w", err)
	}
	slog.Debug("fetched repo",
		"owner", owner, "repo", name,
		"languages", fetched.Metadata.Languages,
		"topics", fetched.Metadata.Topics,
		"downloaded", fetched.Manifest.Downloaded,
		"missing", fetched.Manifest.Missing,
		"dir", fetched.Manifest.Dir,
	)

	deps, err := analyzer.DetectDependencies(fetched, manifestMap)
	if err != nil {
		return model.AnalysisResult{}, nil, fmt.Errorf("detect dependencies: %w", err)
	}
	slog.Debug("detected dependencies", "owner", owner, "repo", name, "count", len(deps), "dependencies", deps)

	result, err := classifier.Classify(ctx, deps, allowed)
	if err != nil {
		return model.AnalysisResult{}, nil, fmt.Errorf("classify: %w", err)
	}
	return result, fetched.Metadata.Languages, nil
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
