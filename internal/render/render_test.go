package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/model"
	"github.com/okarahan/repo-compliance-checker/internal/render"
)

func sampleReport() model.ComplianceReport {
	return model.ComplianceReport{
		Repo:        "owner/repo",
		GeneratedAt: "2026-05-30T14:59:30Z",
		Detected: []model.ReportedTechnology{
			{Name: "Go", Category: model.CategoryLanguage, Allowed: true, Confidence: 1, Bytes: 49116},
			{Name: "Lip Gloss", Category: model.CategoryFramework, Allowed: false, Confidence: 0.85,
				Evidence: []model.Evidence{{File: "go.mod", Snippet: "github.com/charmbracelet/lipgloss v1.1.0"}}},
		},
		Uncertainties: []string{"some note"},
		Conclusion: model.Conclusion{
			DetectedCount: 2,
			AllowedCount:  1,
			Categories: model.CategoryBreakdown{
				Language:  model.CategoryCompliance{DetectedCount: 1, AllowedCount: 1, CompliancePercentage: 92.9},
				Framework: model.CategoryCompliance{DetectedCount: 1, AllowedCount: 0, CompliancePercentage: 0},
			},
			OverallCompliancePercentage: 46.5,
			Compliant:                   false,
		},
	}
}

func TestHTML_containsKeyContent(t *testing.T) {
	html, err := render.HTML(sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	s := string(html)
	for _, want := range []string{
		"<!DOCTYPE html>", "owner/repo", "46.5%", "NON-COMPLIANT",
		"Go", "Lip Gloss", "not allowed", "49116 bytes", "some note",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("rendered HTML missing %q", want)
		}
	}
}

func TestHTML_escapesContent(t *testing.T) {
	r := sampleReport()
	r.Uncertainties = []string{"<script>alert(1)</script>"}
	html, err := render.HTML(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(html), "<script>alert(1)</script>") {
		t.Fatalf("script content should be HTML-escaped")
	}
}

func TestWrite_createsHTMLFile(t *testing.T) {
	dir := t.TempDir()
	path, err := render.Write(dir, sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	if got := filepath.Base(path); got != "owner__repo.html" {
		t.Fatalf("filename=%q", got)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("html file not written: %v", err)
	}
}

func TestRenderJSONFile_writesSiblingHTML(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "owner__repo.json")
	if err := os.WriteFile(jsonPath, []byte(`{"repo":"owner/repo","conclusion":{"overall_compliance_percentage":46.5}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	htmlPath, err := render.RenderJSONFile(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := filepath.Base(htmlPath); got != "owner__repo.html" {
		t.Fatalf("html path=%q", got)
	}
	b, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "owner/repo") {
		t.Fatalf("rendered html missing repo name")
	}
}
