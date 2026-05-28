package endpoints_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/endpoints"
)

func TestGetTopics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/golang/go" {
			t.Fatalf("path=%q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer t" {
			t.Fatalf("Authorization=%q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"topics":["go","compiler"]}`))
	}))
	t.Cleanup(srv.Close)

	c, err := client.New("t", client.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	topics, err := endpoints.GetTopics(context.Background(), c, "golang", "go")
	if err != nil {
		t.Fatal(err)
	}
	if len(topics) != 2 || topics[0] != "go" || topics[1] != "compiler" {
		t.Fatalf("topics=%v", topics)
	}
}

