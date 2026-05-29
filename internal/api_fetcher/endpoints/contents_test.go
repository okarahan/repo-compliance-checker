package endpoints_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/endpoints"
)

func TestGetFileContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/golang/go/contents/go.mod" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		if got := r.URL.Query().Get("ref"); got != "master" {
			t.Fatalf("ref=%q, want master", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":     "file",
			"encoding": "base64",
			"content":  base64.StdEncoding.EncodeToString([]byte("module example\n")),
			"path":     "go.mod",
		})
	}))
	t.Cleanup(srv.Close)

	c, err := client.New("t", client.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	b, err := endpoints.GetFileContent(context.Background(), c, "golang", "go", "master", "go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "module example\n" {
		t.Fatalf("content=%q", string(b))
	}
}

func TestGetFileContent_notFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	c, err := client.New("t", client.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	_, err = endpoints.GetFileContent(context.Background(), c, "golang", "go", "master", "missing.txt")
	if !errors.Is(err, endpoints.ErrFileNotFound) {
		t.Fatalf("err=%v, want ErrFileNotFound", err)
	}
}
