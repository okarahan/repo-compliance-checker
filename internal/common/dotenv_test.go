package common_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/okarahan/repo-compliance-checker/internal/common"
)

func TestReadEnvValue_fromFile(t *testing.T) {
	t.Setenv("RCC_TEST_KEY", "")

	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("RCC_TEST_KEY=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := common.ReadEnvValue("RCC_TEST_KEY", envPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != "from-file" {
		t.Fatalf("value=%q", got)
	}
}

func TestReadEnvValue_envTakesPrecedence(t *testing.T) {
	t.Setenv("RCC_TEST_KEY", "from-env")

	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("RCC_TEST_KEY=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := common.ReadEnvValue("RCC_TEST_KEY", envPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != "from-env" {
		t.Fatalf("value=%q", got)
	}
}

func TestReadEnvValue_missing(t *testing.T) {
	t.Setenv("RCC_TEST_KEY", "")

	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, []byte("# empty\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := common.ReadEnvValue("RCC_TEST_KEY", envPath); err == nil {
		t.Fatal("expected error")
	}
}
