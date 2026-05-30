package report_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/model"
	"github.com/okarahan/repo-compliance-checker/internal/report"
)

func TestBuild_flagsAllowedAndConcludes(t *testing.T) {
	result := model.AnalysisResult{
		Technologies: []model.DetectedTechnology{
			{Name: "Go", Category: model.CategoryLanguage, Confidence: 0.99},
			{Name: "Echo", Category: model.CategoryFramework, Confidence: 0.9},
			{Name: "Cloudflare Go", Category: model.CategoryFramework, Confidence: 0.8},
		},
		Uncertainties: []string{"note"},
	}
	allowed := model.AllowedTechnologies{
		ProgrammingLanguages: []string{"Go"},
		Frameworks:           []string{"Echo"},
	}

	rep := report.Build("owner/repo", result, allowed)

	if len(rep.Detected) != 3 {
		t.Fatalf("detected=%d", len(rep.Detected))
	}
	wantAllowed := map[string]bool{"Go": true, "Echo": true, "Cloudflare Go": false}
	for _, d := range rep.Detected {
		if d.Allowed != wantAllowed[d.Name] {
			t.Fatalf("tech %q allowed=%v, want %v", d.Name, d.Allowed, wantAllowed[d.Name])
		}
	}

	c := rep.Conclusion
	if c.DetectedCount != 3 || c.AllowedCount != 2 {
		t.Fatalf("counts detected=%d allowed=%d", c.DetectedCount, c.AllowedCount)
	}
	if c.AllowedPercentage != 66.7 {
		t.Fatalf("percentage=%v want 66.7", c.AllowedPercentage)
	}
	if c.Compliant {
		t.Fatalf("expected non-compliant when not all allowed")
	}
}

func TestBuild_allAllowedIsCompliant(t *testing.T) {
	result := model.AnalysisResult{
		Technologies: []model.DetectedTechnology{
			{Name: "Go", Category: model.CategoryLanguage},
		},
	}
	allowed := model.AllowedTechnologies{ProgrammingLanguages: []string{"Go"}}

	rep := report.Build("o/r", result, allowed)
	if !rep.Conclusion.Compliant || rep.Conclusion.AllowedPercentage != 100 {
		t.Fatalf("conclusion=%+v", rep.Conclusion)
	}
}

func TestBuild_noTechnologiesIsCompliant(t *testing.T) {
	rep := report.Build("o/r", model.AnalysisResult{}, model.AllowedTechnologies{})
	if !rep.Conclusion.Compliant || rep.Conclusion.AllowedPercentage != 100 {
		t.Fatalf("conclusion=%+v", rep.Conclusion)
	}
}

func TestWrite_createsJSONFile(t *testing.T) {
	dir := t.TempDir()
	rep := model.ComplianceReport{Repo: "owner/repo"}

	path, err := report.Write(dir, rep)
	if err != nil {
		t.Fatal(err)
	}
	if got := filepath.Base(path); got != "owner__repo.json" {
		t.Fatalf("filename=%q", got)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var roundtrip model.ComplianceReport
	if err := json.Unmarshal(b, &roundtrip); err != nil {
		t.Fatalf("written file is not valid JSON: %v", err)
	}
	if roundtrip.Repo != "owner/repo" {
		t.Fatalf("repo=%q", roundtrip.Repo)
	}
}
