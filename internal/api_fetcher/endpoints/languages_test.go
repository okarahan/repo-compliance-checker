package endpoints_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/endpoints"
)

func TestGetLanguages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/golang/go/languages" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer t" {
			t.Fatalf("Authorization=%q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Go":123,"Assembly":4}`))
	}))
	t.Cleanup(srv.Close)

	c, err := client.New("t", client.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	langs, err := endpoints.GetLanguages(context.Background(), c, "golang", "go")
	if err != nil {
		t.Fatal(err)
	}
	if langs["Go"] != 123 {
		t.Fatalf("Go=%d, want 123", langs["Go"])
	}
	if langs["Assembly"] != 4 {
		t.Fatalf("Assembly=%d, want 4", langs["Assembly"])
	}
}

