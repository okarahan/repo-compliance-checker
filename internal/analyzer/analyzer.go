package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher"
	"github.com/okarahan/repo-compliance-checker/internal/model"
)

// languageDependencyKeywords maps a (lowercase) programming language to the keywords
// that mark a dependency declaration inside its manifest files.
//
// Example: in a Go `go.mod` a dependency line is introduced by `require`, in a
// `package.json` the relevant sections are `dependencies` / `devDependencies`.
//
// An empty keyword list means "every non-empty, non-comment line counts as a candidate"
// (e.g. Python `requirements.txt`, where each line is a dependency).
var languageDependencyKeywords = map[string][]string{
	"go":         {"require"},
	"java":       {"dependency", "implementation", "testimplementation", "api", "compile"},
	"kotlin":     {"dependency", "implementation", "testimplementation", "api", "compile"},
	"javascript": {"dependencies", "devdependencies"},
	"typescript": {"dependencies", "devdependencies"},
	"python":     {},
}

// DetectDependencies reads the manifest files that were downloaded for a repo and
// extracts the raw dependency lines per detected language.
//
// This is the first (deterministic) analysis layer: it only collects candidate lines
// as evidence. Mapping those lines to concrete technologies happens in a later step.
func DetectDependencies(result api_fetcher.RepoFetchResult, manifestCfg model.ManifestMapConfig) ([]model.RawDependency, error) {
	downloaded := make(map[string]struct{}, len(result.Manifest.Downloaded))
	for _, p := range result.Manifest.Downloaded {
		downloaded[p] = struct{}{}
	}

	var findings []model.RawDependency
	processed := make(map[string]struct{})

	for lang := range result.Metadata.Languages {
		key := strings.ToLower(strings.TrimSpace(lang))
		targets, ok := manifestCfg.ByLanguage[key]
		if !ok {
			continue
		}
		keywords := languageDependencyKeywords[key]

		for _, file := range targets.Files {
			if _, ok := downloaded[file]; !ok {
				continue
			}
			if _, done := processed[file]; done {
				continue
			}
			processed[file] = struct{}{}

			lines, err := scanFile(filepath.Join(result.Manifest.Dir, filepath.FromSlash(file)), key, file, keywords)
			if err != nil {
				return nil, err
			}
			for _, line := range lines {
				name := normalizeName(key, line)
				if name == "" {
					continue
				}
				findings = append(findings, model.RawDependency{
					Language: key,
					File:     file,
					Line:     line,
					Name:     name,
				})
			}
		}
	}

	return findings, nil
}

// scanFile returns the trimmed candidate lines of a manifest file.
//
// It is block-aware for Go: dependencies inside a `require ( ... )` block span
// multiple lines that do not themselves contain the `require` keyword, so we track
// whether we are currently inside such a block. Outside of blocks (and for other
// languages) it falls back to keyword matching; an empty keyword list means every
// non-empty, non-comment line is a candidate.
//
// Maven `pom.xml` is handled separately, because its dependencies are multi-line
// XML blocks (`<dependency>...</dependency>`) whose identifier is split across the
// `<groupId>` and `<artifactId>` child elements.
func scanFile(path, language, file string, keywords []string) ([]string, error) {
	if file == "pom.xml" {
		return scanMavenPom(path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open manifest %q: %w", path, err)
	}
	defer f.Close()

	var out []string
	inGoRequireBlock := false

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || isComment(line) {
			continue
		}

		if language == "go" {
			if inGoRequireBlock {
				if line == ")" {
					inGoRequireBlock = false
					continue
				}
				out = append(out, line)
				continue
			}
			if isGoRequireBlockOpen(line) {
				inGoRequireBlock = true
				continue
			}
		}

		if matchesKeyword(line, keywords) {
			out = append(out, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read manifest %q: %w", path, err)
	}
	return out, nil
}

// isGoRequireBlockOpen reports whether the line opens a `require ( ... )` block,
// e.g. "require (".
func isGoRequireBlockOpen(line string) bool {
	if !strings.HasPrefix(line, "require") {
		return false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, "require"))
	return rest == "("
}

// scanMavenPom extracts the dependencies from a Maven pom.xml. Each
// `<dependency>...</dependency>` block is collapsed into a single
// "groupId:artifactId" identifier (falling back to whichever part is present).
func scanMavenPom(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open manifest %q: %w", path, err)
	}
	defer f.Close()

	var out []string
	inDependency := false
	var groupID, artifactID string

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if strings.Contains(line, "<dependency>") {
			inDependency = true
			groupID, artifactID = "", ""
			continue
		}
		if !inDependency {
			continue
		}
		if strings.Contains(line, "</dependency>") {
			inDependency = false
			if name := mavenName(groupID, artifactID); name != "" {
				out = append(out, name)
			}
			continue
		}
		if v, ok := xmlTagValue(line, "groupId"); ok {
			groupID = v
		}
		if v, ok := xmlTagValue(line, "artifactId"); ok {
			artifactID = v
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read manifest %q: %w", path, err)
	}
	return out, nil
}

// mavenName joins a Maven groupId and artifactId into "groupId:artifactId",
// falling back to whichever part is present.
func mavenName(groupID, artifactID string) string {
	groupID = strings.TrimSpace(groupID)
	artifactID = strings.TrimSpace(artifactID)
	switch {
	case groupID != "" && artifactID != "":
		return groupID + ":" + artifactID
	case artifactID != "":
		return artifactID
	default:
		return groupID
	}
}

// xmlTagValue returns the text content of a single-line XML element, e.g.
// xmlTagValue("<groupId>org.springframework</groupId>", "groupId") -> "org.springframework".
func xmlTagValue(line, tag string) (string, bool) {
	open, closing := "<"+tag+">", "</"+tag+">"
	i := strings.Index(line, open)
	if i < 0 {
		return "", false
	}
	j := strings.Index(line, closing)
	if j < i+len(open) {
		return "", false
	}
	return strings.TrimSpace(line[i+len(open) : j]), true
}

func matchesKeyword(line string, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}
	lower := strings.ToLower(line)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func isComment(line string) bool {
	return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//")
}

// normalizeName strips the version off a matched manifest line and returns the bare
// dependency identifier. Version notation is format-specific, so this is per language.
// For unsupported languages it returns the line unchanged.
func normalizeName(language, line string) string {
	switch language {
	case "go":
		return normalizeGo(line)
	case "python":
		return normalizePython(line)
	default:
		return strings.TrimSpace(line)
	}
}

// normalizeGo turns e.g. "require github.com/labstack/echo/v4 v4.11.0 // indirect"
// into "github.com/labstack/echo/v4". Block markers like "require (" yield "".
func normalizeGo(line string) string {
	s := strings.TrimSpace(line)
	s = strings.TrimSpace(strings.TrimPrefix(s, "require"))
	if i := strings.Index(s, "//"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if s == "" || s == "(" {
		return ""
	}
	fields := strings.Fields(s)
	return fields[0]
}

// normalizePython turns e.g. "testcontainers[postgres]==4.0  # comment" into "testcontainers".
func normalizePython(line string) string {
	s := strings.TrimSpace(line)
	if i := strings.Index(s, "#"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if i := strings.Index(s, ";"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}

	cut := len(s)
	for _, op := range []string{"===", "==", ">=", "<=", "~=", "!=", ">", "<", "="} {
		if i := strings.Index(s, op); i >= 0 && i < cut {
			cut = i
		}
	}
	s = strings.TrimSpace(s[:cut])

	if i := strings.Index(s, "["); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	return s
}
