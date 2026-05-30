// Package report turns an analyzer AnalysisResult into a compliance report:
// it matches detected technologies against the allow-list, derives an aggregate
// conclusion, and writes the result as a JSON file to the local filesystem.
package report

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/okarahan/repo-compliance-checker/internal/model"
)

// Build assembles a ComplianceReport from the analysis result for a repo. It
// combines the GitHub Linguist languages (byte-weighted) with the technologies
// classified from manifest dependencies, flags each as allowed/not-allowed, and
// computes the aggregate conclusion.
func Build(slug string, result model.AnalysisResult, languages map[string]int64, allowed model.AllowedTechnologies) model.ComplianceReport {
	allowSet := allowedSet(allowed)

	detected := make([]model.ReportedTechnology, 0, len(languages)+len(result.Technologies))

	// Languages come from GitHub Linguist, ordered by code size (descending).
	for _, ls := range sortedLanguages(languages) {
		detected = append(detected, model.ReportedTechnology{
			Name:       ls.name,
			Category:   model.CategoryLanguage,
			Allowed:    allowSet[normalize(ls.name)],
			Confidence: 1,
			Bytes:      ls.bytes,
			Evidence:   []model.Evidence{{File: "github linguist", Snippet: fmt.Sprintf("%d bytes", ls.bytes)}},
		})
	}

	// Remaining technologies come from the dependency classification.
	for _, t := range result.Technologies {
		detected = append(detected, model.ReportedTechnology{
			Name:       t.Name,
			Category:   t.Category,
			Allowed:    allowSet[normalize(t.Name)],
			Confidence: t.Confidence,
			Evidence:   t.Evidence,
		})
	}

	allowedCount := 0
	for _, d := range detected {
		if d.Allowed {
			allowedCount++
		}
	}

	return model.ComplianceReport{
		Repo:          slug,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Detected:      detected,
		Uncertainties: result.Uncertainties,
		Conclusion:    conclude(detected, languages, allowSet, allowedCount),
	}
}

// langStat is a language name paired with its Linguist byte count.
type langStat struct {
	name  string
	bytes int64
}

// sortedLanguages returns the languages ordered by byte count (descending), then
// by name, for deterministic report output.
func sortedLanguages(languages map[string]int64) []langStat {
	stats := make([]langStat, 0, len(languages))
	for name, bytes := range languages {
		if strings.TrimSpace(name) == "" {
			continue
		}
		stats = append(stats, langStat{name: name, bytes: bytes})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].bytes != stats[j].bytes {
			return stats[i].bytes > stats[j].bytes
		}
		return stats[i].name < stats[j].name
	})
	return stats
}

// Category weights for the overall compliance score. They sum to 1.0.
const (
	weightLanguage  = 0.5
	weightFramework = 0.3
	weightUtility   = 0.2
)

// conclude derives the per-category compliance and the weighted overall score.
// Languages are byte-weighted (Linguist code size); frameworks and utilities are
// fractions of allowed technologies. A category with nothing detected counts as
// 0%. The repo is compliant only when the weighted overall compliance is 100%.
func conclude(detected []model.ReportedTechnology, languages map[string]int64, allowSet map[string]bool, allowedCount int) model.Conclusion {
	lang := languageCompliance(languages, allowSet)
	framework := categoryCompliance(detected, model.CategoryFramework)
	util := categoryCompliance(detected, model.CategoryUtility)

	overall := roundTo1(
		weightLanguage*lang.CompliancePercentage +
			weightFramework*framework.CompliancePercentage +
			weightUtility*util.CompliancePercentage,
	)

	return model.Conclusion{
		DetectedCount: len(detected),
		AllowedCount:  allowedCount,
		Categories: model.CategoryBreakdown{
			Language:  lang,
			Framework: framework,
			Utility:   util,
		},
		OverallCompliancePercentage: overall,
		Compliant:                   overall == 100,
	}
}

// languageCompliance computes the language compliance weighted by code size:
// the share of bytes belonging to allowed languages over the total bytes. With no
// language bytes it is treated as 0%.
func languageCompliance(languages map[string]int64, allowSet map[string]bool) model.CategoryCompliance {
	var totalBytes, allowedBytes int64
	var detectedCount, allowedCount int
	for name, bytes := range languages {
		if strings.TrimSpace(name) == "" {
			continue
		}
		detectedCount++
		totalBytes += bytes
		if allowSet[normalize(name)] {
			allowedCount++
			allowedBytes += bytes
		}
	}

	percentage := 0.0
	if totalBytes > 0 {
		percentage = roundTo1(float64(allowedBytes) / float64(totalBytes) * 100)
	}
	return model.CategoryCompliance{
		DetectedCount:        detectedCount,
		AllowedCount:         allowedCount,
		CompliancePercentage: percentage,
	}
}

// categoryCompliance computes the compliance for a single category. A category
// with no detected technologies is treated as 0% compliant.
func categoryCompliance(detected []model.ReportedTechnology, category model.Category) model.CategoryCompliance {
	var detectedCount, allowedCount int
	for _, t := range detected {
		if t.Category != category {
			continue
		}
		detectedCount++
		if t.Allowed {
			allowedCount++
		}
	}

	percentage := 0.0
	if detectedCount > 0 {
		percentage = roundTo1(float64(allowedCount) / float64(detectedCount) * 100)
	}
	return model.CategoryCompliance{
		DetectedCount:        detectedCount,
		AllowedCount:         allowedCount,
		CompliancePercentage: percentage,
	}
}

// Write serializes the report as pretty JSON into dir, creating dir if needed,
// and returns the path of the written file.
func Write(dir string, r model.ComplianceReport) (string, error) {
	if strings.TrimSpace(dir) == "" {
		dir = "reports"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create report dir: %w", err)
	}

	path := filepath.Join(dir, fileName(r.Repo))
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return "", fmt.Errorf("write report %q: %w", path, err)
	}
	return path, nil
}

// fileName builds a filesystem-safe report filename from a repo slug, e.g.
// "owner/repo" -> "owner__repo.json".
func fileName(slug string) string {
	safe := strings.ReplaceAll(strings.TrimSpace(slug), "/", "__")
	if safe == "" {
		safe = "report"
	}
	return safe + ".json"
}

// allowedSet builds a lookup set of normalized allowed technology names across
// all allow-list categories.
func allowedSet(allowed model.AllowedTechnologies) map[string]bool {
	set := map[string]bool{}
	for _, group := range [][]string{
		allowed.ProgrammingLanguages,
		allowed.Frameworks,
		allowed.Utilities,
	} {
		for _, name := range group {
			if n := normalize(name); n != "" {
				set[n] = true
			}
		}
	}
	return set
}

// normalize lower-cases a technology name and strips every non-alphanumeric
// character so that variants like "Node.js", "node js" and "nodejs" compare
// equal.
func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func roundTo1(v float64) float64 {
	return math.Round(v*10) / 10
}
