package client_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/api_fetcher/client"
)

func TestNew_requiresToken(t *testing.T) {
	_, err := client.New("", client.Options{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_NewRequest_setsAuthHeader(t *testing.T) {
	c, err := client.New("abc123", client.Options{BaseURL: "https://example.test/"})
	if err != nil {
		t.Fatal(err)
	}

	req, err := c.NewRequest(http.MethodGet, "/repos/org/repo/languages")
	if err != nil {
		t.Fatal(err)
	}

	if got := req.Header.Get("Authorization"); got != "Bearer abc123" {
		t.Fatalf("Authorization=%q, want %q", got, "Bearer abc123")
	}
	if got := req.Header.Get("Accept"); got == "" {
		t.Fatalf("Accept header should be set")
	}
	if got := req.Header.Get("User-Agent"); got == "" {
		t.Fatalf("User-Agent header should be set")
	}
}

func TestClient_Do_usesHTTPClient(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c, err := client.New("t", client.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	req, err := c.NewRequest(http.MethodGet, "/x")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if gotAuth != "Bearer t" {
		t.Fatalf("server saw Authorization=%q, want %q", gotAuth, "Bearer t")
	}
}

