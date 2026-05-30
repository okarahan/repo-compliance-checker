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
			{Name: "Echo", Category: model.CategoryFramework, Confidence: 0.9},
			{Name: "Spring", Category: model.CategoryFramework, Confidence: 0.8},
			{Name: "Docker", Category: model.CategoryUtility, Confidence: 0.8},
		},
		Uncertainties: []string{"note"},
	}
	languages := map[string]int64{"Go": 9000, "HTML": 1000}
	allowed := model.AllowedTechnologies{
		ProgrammingLanguages: []string{"Go"},
		Frameworks:           []string{"Echo"},
		Utilities:            []string{"Docker"},
	}

	rep := report.Build("owner/repo", result, languages, allowed)

	// 2 languages + 3 classified technologies.
	if len(rep.Detected) != 5 {
		t.Fatalf("detected=%d", len(rep.Detected))
	}
	wantAllowed := map[string]bool{"Go": true, "HTML": false, "Echo": true, "Spring": false, "Docker": true}
	for _, d := range rep.Detected {
		if d.Allowed != wantAllowed[d.Name] {
			t.Fatalf("tech %q allowed=%v, want %v", d.Name, d.Allowed, wantAllowed[d.Name])
		}
	}

	c := rep.Conclusion
	if c.DetectedCount != 5 || c.AllowedCount != 3 {
		t.Fatalf("counts detected=%d allowed=%d", c.DetectedCount, c.AllowedCount)
	}
	// language is byte-weighted: 9000 of 10000 bytes are allowed -> 90%.
	if c.Categories.Language.CompliancePercentage != 90 {
		t.Fatalf("language=%v want 90", c.Categories.Language.CompliancePercentage)
	}
	if c.Categories.Framework.CompliancePercentage != 50 {
		t.Fatalf("framework=%v want 50", c.Categories.Framework.CompliancePercentage)
	}
	if c.Categories.Utility.CompliancePercentage != 100 {
		t.Fatalf("utility=%v want 100", c.Categories.Utility.CompliancePercentage)
	}
	// overall = 0.5*90 + 0.3*50 + 0.2*100 = 80
	if c.OverallCompliancePercentage != 80 {
		t.Fatalf("overall=%v want 80", c.OverallCompliancePercentage)
	}
	if c.Compliant {
		t.Fatalf("expected non-compliant when overall < 100")
	}
}

func TestBuild_languageComplianceIsByteWeighted(t *testing.T) {
	languages := map[string]int64{"Go": 49116, "HTML": 1487, "Makefile": 1039, "Shell": 1236}
	allowed := model.AllowedTechnologies{ProgrammingLanguages: []string{"Go"}}

	rep := report.Build("o/r", model.AnalysisResult{}, languages, allowed)
	lang := rep.Conclusion.Categories.Language
	if lang.DetectedCount != 4 || lang.AllowedCount != 1 {
		t.Fatalf("language counts detected=%d allowed=%d", lang.DetectedCount, lang.AllowedCount)
	}
	// 49116 / 52878 = 92.9%
	if lang.CompliancePercentage != 92.9 {
		t.Fatalf("language=%v want 92.9", lang.CompliancePercentage)
	}
	// languages are listed biggest-first, with byte counts.
	if rep.Detected[0].Name != "Go" || rep.Detected[0].Bytes != 49116 {
		t.Fatalf("first detected=%+v", rep.Detected[0])
	}
}

func TestBuild_allCategoriesAllowedIsCompliant(t *testing.T) {
	result := model.AnalysisResult{
		Technologies: []model.DetectedTechnology{
			{Name: "Echo", Category: model.CategoryFramework},
			{Name: "Docker", Category: model.CategoryUtility},
		},
	}
	languages := map[string]int64{"Go": 100}
	allowed := model.AllowedTechnologies{
		ProgrammingLanguages: []string{"Go"},
		Frameworks:           []string{"Echo"},
		Utilities:            []string{"Docker"},
	}

	rep := report.Build("o/r", result, languages, allowed)
	if !rep.Conclusion.Compliant || rep.Conclusion.OverallCompliancePercentage != 100 {
		t.Fatalf("conclusion=%+v", rep.Conclusion)
	}
}

func TestBuild_emptyCategoryCountsAsZero(t *testing.T) {
	languages := map[string]int64{"Go": 100}
	allowed := model.AllowedTechnologies{ProgrammingLanguages: []string{"Go"}}

	rep := report.Build("o/r", model.AnalysisResult{}, languages, allowed)
	c := rep.Conclusion
	if c.Categories.Language.CompliancePercentage != 100 {
		t.Fatalf("language=%v want 100", c.Categories.Language.CompliancePercentage)
	}
	if c.Categories.Framework.CompliancePercentage != 0 || c.Categories.Utility.CompliancePercentage != 0 {
		t.Fatalf("empty categories should be 0%%, got framework=%v util=%v",
			c.Categories.Framework.CompliancePercentage, c.Categories.Utility.CompliancePercentage)
	}
	// overall = 0.5*100 + 0.3*0 + 0.2*0 = 50
	if c.OverallCompliancePercentage != 50 {
		t.Fatalf("overall=%v want 50", c.OverallCompliancePercentage)
	}
	if c.Compliant {
		t.Fatalf("expected non-compliant when overall < 100")
	}
}

func TestBuild_noTechnologiesIsNotCompliant(t *testing.T) {
	rep := report.Build("o/r", model.AnalysisResult{}, nil, model.AllowedTechnologies{})
	if rep.Conclusion.Compliant || rep.Conclusion.OverallCompliancePercentage != 0 {
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
