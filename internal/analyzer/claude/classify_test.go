package claude_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/analyzer/claude"
	"github.com/okarahan/repo-compliance-checker/internal/model"
)

func TestClassify(t *testing.T) {
	var sawAPIKey, sawVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" || r.Method != http.MethodPost {
			t.Fatalf("path=%q method=%q", r.URL.Path, r.Method)
		}
		sawAPIKey = r.Header.Get("x-api-key")
		sawVersion = r.Header.Get("anthropic-version")
		_, _ = io.ReadAll(r.Body)

		modelJSON := `{"technologies":[{"name":"Echo","category":"framework","source_dependency":"github.com/labstack/echo/v4","confidence":0.95}],"uncertainties":["could not classify foobar"]}`
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{{"type": "text", "text": modelJSON}},
		})
	}))
	t.Cleanup(srv.Close)

	c, err := claude.New("secret-key", claude.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	deps := []model.RawDependency{
		{Language: "go", File: "go.mod", Line: "require github.com/labstack/echo/v4 v4.11.0", Name: "github.com/labstack/echo/v4"},
	}
	allowed := model.AllowedTechnologies{Frameworks: []string{"Echo", "Spring"}}

	res, err := c.Classify(context.Background(), deps, allowed)
	if err != nil {
		t.Fatal(err)
	}

	if sawAPIKey != "secret-key" {
		t.Fatalf("x-api-key=%q", sawAPIKey)
	}
	if sawVersion == "" {
		t.Fatalf("anthropic-version header should be set")
	}
	if len(res.Technologies) != 1 {
		t.Fatalf("technologies=%+v", res.Technologies)
	}
	got := res.Technologies[0]
	if got.Name != "Echo" || got.Category != model.CategoryFramework {
		t.Fatalf("tech=%+v", got)
	}
	if len(got.Evidence) != 1 || got.Evidence[0].File != "go.mod" {
		t.Fatalf("evidence=%+v", got.Evidence)
	}
	if len(res.Uncertainties) != 1 {
		t.Fatalf("uncertainties=%+v", res.Uncertainties)
	}
}

func TestClassify_noDepsSkipsCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called when there are no dependencies")
	}))
	t.Cleanup(srv.Close)

	c, err := claude.New("k", claude.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.Classify(context.Background(), nil, model.AllowedTechnologies{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Technologies) != 0 {
		t.Fatalf("expected empty result, got %+v", res)
	}
}
