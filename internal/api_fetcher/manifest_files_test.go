package api_fetcher_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/endpoints"
	"github.com/okarahan/repo-compliance-checker/internal/model"
)

func TestResolveManifestPaths(t *testing.T) {
	langs := endpoints.LanguagesResponse{"Go": 100, "JavaScript": 50}
	mm := model.ManifestMapConfig{
		ByLanguage: map[string]model.ManifestTargets{
			"go":         {Files: []string{"go.mod", "go.sum"}},
			"javascript": {Files: []string{"package.json"}},
			"python":     {Files: []string{"requirements.txt"}},
		},
		Global: model.ManifestTargets{Files: []string{"Dockerfile"}},
	}

	got := api_fetcher.ResolveManifestPaths(langs, mm)
	sort.Strings(got)

	want := []string{"Dockerfile", "go.mod", "go.sum", "package.json"}
	if len(got) != len(want) {
		t.Fatalf("got=%v, want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got=%v, want=%v", got, want)
		}
	}
}

func TestDownloadManifestFiles(t *testing.T) {
	files := map[string]string{
		"go.mod":     "module example\n",
		"Dockerfile": "FROM golang:1.22\n",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /repos/golang/go/contents/<path>
		const prefix = "/repos/golang/go/contents/"
		path := r.URL.Path[len(prefix):]
		content, ok := files[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":     "file",
			"encoding": "base64",
			"content":  base64.StdEncoding.EncodeToString([]byte(content)),
			"path":     path,
		})
	}))
	t.Cleanup(srv.Close)

	c, err := client.New("t", client.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	langs := endpoints.LanguagesResponse{"Go": 100}
	mm := model.ManifestMapConfig{
		ByLanguage: map[string]model.ManifestTargets{
			"go": {Files: []string{"go.mod", "go.sum"}},
		},
		Global: model.ManifestTargets{Files: []string{"Dockerfile"}},
	}

	dest := t.TempDir()
	res, err := api_fetcher.DownloadManifestFiles(context.Background(), c, "golang", "go", "master", langs, mm, dest)
	if err != nil {
		t.Fatal(err)
	}

	sort.Strings(res.Downloaded)
	wantDownloaded := []string{"Dockerfile", "go.mod"}
	if len(res.Downloaded) != len(wantDownloaded) {
		t.Fatalf("downloaded=%v, want=%v", res.Downloaded, wantDownloaded)
	}
	for i := range wantDownloaded {
		if res.Downloaded[i] != wantDownloaded[i] {
			t.Fatalf("downloaded=%v, want=%v", res.Downloaded, wantDownloaded)
		}
	}

	if len(res.Missing) != 1 || res.Missing[0] != "go.sum" {
		t.Fatalf("missing=%v, want [go.sum]", res.Missing)
	}

	b, err := os.ReadFile(filepath.Join(dest, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "module example\n" {
		t.Fatalf("go.mod content=%q", string(b))
	}
}
