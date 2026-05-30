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
	"strings"
	"time"

	"github.com/okarahan/repo-compliance-checker/internal/model"
)

// Build assembles a ComplianceReport from the analysis result for a repo,
// flagging each detected technology as allowed/not-allowed and computing the
// aggregate conclusion.
func Build(slug string, result model.AnalysisResult, allowed model.AllowedTechnologies) model.ComplianceReport {
	allowSet := allowedSet(allowed)

	detected := make([]model.ReportedTechnology, 0, len(result.Technologies))
	allowedCount := 0
	for _, t := range result.Technologies {
		isAllowed := allowSet[normalize(t.Name)]
		if isAllowed {
			allowedCount++
		}
		detected = append(detected, model.ReportedTechnology{
			Name:       t.Name,
			Category:   t.Category,
			Allowed:    isAllowed,
			Confidence: t.Confidence,
			Evidence:   t.Evidence,
		})
	}

	return model.ComplianceReport{
		Repo:          slug,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Detected:      detected,
		Uncertainties: result.Uncertainties,
		Conclusion:    conclude(len(detected), allowedCount),
	}
}

// conclude derives the aggregate compliance outcome. A repo is compliant only
// when every detected technology is allowed. With no detected technologies there
// is nothing non-compliant, so it counts as 100% / compliant.
func conclude(detectedCount, allowedCount int) model.Conclusion {
	percentage := 100.0
	if detectedCount > 0 {
		percentage = roundTo1(float64(allowedCount) / float64(detectedCount) * 100)
	}
	return model.Conclusion{
		DetectedCount:     detectedCount,
		AllowedCount:      allowedCount,
		AllowedPercentage: percentage,
		Compliant:         allowedCount == detectedCount,
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
