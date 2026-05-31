// Package render turns a ComplianceReport into a self-contained HTML page so the
// JSON report can be viewed in a browser at a glance.
package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/okarahan/repo-compliance-checker/internal/model"
)

// categoryRow is a single category line in the per-category breakdown.
type categoryRow struct {
	Name       string
	Weight     string
	Compliance model.CategoryCompliance
}

// htmlView is the data passed to the HTML template.
type htmlView struct {
	Report     model.ComplianceReport
	Categories []categoryRow
}

var funcs = template.FuncMap{
	"pf1": func(f float64) string { return fmt.Sprintf("%.1f", f) },
	"pf2": func(f float64) string { return fmt.Sprintf("%.2f", f) },
	// barClass picks a color class from a 0-100 percentage.
	"barClass": func(p float64) string {
		switch {
		case p >= 100:
			return "ok"
		case p >= 50:
			return "warn"
		default:
			return "bad"
		}
	},
}

var page = template.Must(template.New("report").Funcs(funcs).Parse(htmlTemplate))

// HTML renders the report as a standalone HTML document.
func HTML(r model.ComplianceReport) ([]byte, error) {
	view := htmlView{
		Report: r,
		Categories: []categoryRow{
			{Name: "Language", Weight: "50%", Compliance: r.Conclusion.Categories.Language},
			{Name: "Framework", Weight: "30%", Compliance: r.Conclusion.Categories.Framework},
			{Name: "Utility", Weight: "20%", Compliance: r.Conclusion.Categories.Utility},
		},
	}

	var buf bytes.Buffer
	if err := page.Execute(&buf, view); err != nil {
		return nil, fmt.Errorf("render html: %w", err)
	}
	return buf.Bytes(), nil
}

// Write renders the report to HTML and writes it into dir, returning the path of
// the written file. The filename mirrors the JSON report (owner__repo.html).
func Write(dir string, r model.ComplianceReport) (string, error) {
	if strings.TrimSpace(dir) == "" {
		dir = "reports"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create render dir: %w", err)
	}

	html, err := HTML(r)
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, fileName(r.Repo))
	if err := os.WriteFile(path, html, 0o644); err != nil {
		return "", fmt.Errorf("write html %q: %w", path, err)
	}
	return path, nil
}

// RenderJSONFile reads a JSON compliance report from jsonPath, renders it to HTML
// and writes the HTML next to it (same base name, .html extension). It returns the
// path of the written HTML file.
func RenderJSONFile(jsonPath string) (string, error) {
	b, err := os.ReadFile(jsonPath)
	if err != nil {
		return "", fmt.Errorf("read report %q: %w", jsonPath, err)
	}
	var r model.ComplianceReport
	if err := json.Unmarshal(b, &r); err != nil {
		return "", fmt.Errorf("parse report %q: %w", jsonPath, err)
	}

	html, err := HTML(r)
	if err != nil {
		return "", err
	}

	outPath := strings.TrimSuffix(jsonPath, filepath.Ext(jsonPath)) + ".html"
	if err := os.WriteFile(outPath, html, 0o644); err != nil {
		return "", fmt.Errorf("write html %q: %w", outPath, err)
	}
	return outPath, nil
}

// fileName builds an HTML report filename from a repo slug (owner/repo -> owner__repo.html).
func fileName(slug string) string {
	safe := strings.ReplaceAll(strings.TrimSpace(slug), "/", "__")
	if safe == "" {
		safe = "report"
	}
	return safe + ".html"
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Compliance report – {{.Report.Repo}}</title>
<style>
  :root { color-scheme: light dark; }
  * { box-sizing: border-box; }
  body { font-family: -apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif; margin: 0; padding: 2rem; background: #f5f6f8; color: #1c1e21; }
  .wrap { max-width: 960px; margin: 0 auto; }
  h1 { margin: 0 0 .25rem; font-size: 1.5rem; }
  h1 code { background: #e9ebee; padding: .1rem .4rem; border-radius: 6px; }
  .muted { color: #6b7280; font-size: .85rem; margin: 0 0 1.5rem; }
  .card { background: #fff; border: 1px solid #e5e7eb; border-radius: 12px; padding: 1.25rem 1.5rem; margin-bottom: 1.25rem; box-shadow: 0 1px 2px rgba(0,0,0,.04); }
  .overall { display: flex; align-items: center; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
  .overall .score { font-size: 2.5rem; font-weight: 700; }
  .badge { display: inline-block; padding: .3rem .75rem; border-radius: 999px; font-weight: 600; font-size: .85rem; }
  .badge.ok { background: #e6f4ea; color: #137333; }
  .badge.bad { background: #fce8e6; color: #c5221f; }
  h2 { font-size: 1rem; text-transform: uppercase; letter-spacing: .04em; color: #6b7280; margin: 0 0 1rem; }
  .cat { margin-bottom: 1rem; }
  .cat:last-child { margin-bottom: 0; }
  .cat .label { display: flex; justify-content: space-between; font-size: .9rem; margin-bottom: .35rem; }
  .cat .label .name { font-weight: 600; }
  .bar { height: 10px; border-radius: 999px; background: #eceef1; overflow: hidden; }
  .bar > span { display: block; height: 100%; border-radius: 999px; }
  .bar > span.ok { background: #34a853; }
  .bar > span.warn { background: #f9ab00; }
  .bar > span.bad { background: #ea4335; }
  table { width: 100%; border-collapse: collapse; font-size: .9rem; }
  th, td { text-align: left; padding: .5rem .6rem; border-bottom: 1px solid #eceef1; vertical-align: top; }
  th { color: #6b7280; font-weight: 600; }
  .pill { font-size: .75rem; padding: .15rem .5rem; border-radius: 999px; font-weight: 600; }
  .pill.ok { background: #e6f4ea; color: #137333; }
  .pill.bad { background: #fce8e6; color: #c5221f; }
  .cat-tag { font-size: .75rem; color: #6b7280; }
  .evi { color: #6b7280; font-size: .8rem; }
  ul.notes { margin: 0; padding-left: 1.2rem; }
  ul.notes li { margin: .2rem 0; }
</style>
</head>
<body>
<div class="wrap">
  <h1>Compliance report · <code>{{.Report.Repo}}</code></h1>
  <p class="muted">Generated at {{.Report.GeneratedAt}}</p>

  <div class="card overall">
    <div>
      <div class="score">{{pf1 .Report.Conclusion.OverallCompliancePercentage}}%</div>
      <div class="muted">Weighted overall compliance · {{.Report.Conclusion.AllowedCount}}/{{.Report.Conclusion.DetectedCount}} technologies allowed</div>
    </div>
    {{if .Report.Conclusion.Compliant}}
      <span class="badge ok">COMPLIANT</span>
    {{else}}
      <span class="badge bad">NON-COMPLIANT</span>
    {{end}}
  </div>

  <div class="card">
    <h2>Category breakdown</h2>
    {{range .Categories}}
    <div class="cat">
      <div class="label">
        <span class="name">{{.Name}} <span class="cat-tag">(weight {{.Weight}})</span></span>
        <span>{{pf1 .Compliance.CompliancePercentage}}% · {{.Compliance.AllowedCount}}/{{.Compliance.DetectedCount}}</span>
      </div>
      <div class="bar"><span class="{{barClass .Compliance.CompliancePercentage}}" style="width:{{pf1 .Compliance.CompliancePercentage}}%"></span></div>
    </div>
    {{end}}
  </div>

  <div class="card">
    <h2>Detected technologies ({{len .Report.Detected}})</h2>
    <table>
      <thead>
        <tr><th>Name</th><th>Category</th><th>Allowed</th><th>Detail</th><th>Evidence</th></tr>
      </thead>
      <tbody>
        {{range .Report.Detected}}
        <tr>
          <td><strong>{{.Name}}</strong></td>
          <td class="cat-tag">{{.Category}}</td>
          <td>{{if .Allowed}}<span class="pill ok">allowed</span>{{else}}<span class="pill bad">not allowed</span>{{end}}</td>
          <td class="evi">{{if gt .Bytes 0}}{{.Bytes}} bytes{{else}}conf {{pf2 .Confidence}}{{end}}</td>
          <td class="evi">{{range .Evidence}}{{.File}}: {{.Snippet}}<br>{{end}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>

  {{if .Report.Uncertainties}}
  <div class="card">
    <h2>Uncertainties</h2>
    <ul class="notes">
      {{range .Report.Uncertainties}}<li>{{.}}</li>{{end}}
    </ul>
  </div>
  {{end}}
</div>
</body>
</html>
`
