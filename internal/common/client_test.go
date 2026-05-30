package common_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/common"
)

func TestNew_requiresAbsoluteBaseURL(t *testing.T) {
	if _, err := common.New(common.Options{}); err == nil {
		t.Fatal("expected error for empty base url")
	}
	if _, err := common.New(common.Options{BaseURL: "/relative"}); err == nil {
		t.Fatal("expected error for non-absolute base url")
	}
}

func TestClient_NewRequest_setsHeadersAndBody(t *testing.T) {
	c, err := common.New(common.Options{
		BaseURL: "https://example.test/",
		Headers: map[string]string{"X-Api-Key": "secret", "Content-Type": "application/json"},
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := c.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}

	if got := req.Header.Get("X-Api-Key"); got != "secret" {
		t.Fatalf("X-Api-Key=%q", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type=%q", got)
	}
	if req.URL.String() != "https://example.test/v1/messages" {
		t.Fatalf("url=%q", req.URL.String())
	}
}

func TestClient_Do(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"a":1}` {
			t.Errorf("body=%q", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c, err := common.New(common.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}

	req, err := c.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}
