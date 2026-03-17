package script_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kanywst/galick/pkg/protocols/script"
)

func writeStarFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.star")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write star file: %v", err)
	}
	return path
}

func TestScriptAttacker_WithHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "hello" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Header.Get("Authorization") != "Bearer abc" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	starScript := `
def request():
    return {
        "method": "GET",
        "url": "` + srv.URL + `",
        "headers": {
            "X-Custom": "hello",
            "Authorization": "Bearer abc",
        },
    }
`
	path := writeStarFile(t, starScript)
	a, err := script.NewScriptAttacker(path, 5*time.Second, false)
	if err != nil {
		t.Fatalf("failed to create attacker: %v", err)
	}

	res := a.Attack(context.Background())

	if res.Code != 200 {
		t.Errorf("expected 200, got %d (headers not applied)", res.Code)
	}
	if res.Error != "" {
		t.Errorf("unexpected error: %s", res.Error)
	}
}

func TestScriptAttacker_NoHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	starScript := `
def request():
    return {
        "method": "GET",
        "url": "` + srv.URL + `",
    }
`
	path := writeStarFile(t, starScript)
	a, err := script.NewScriptAttacker(path, 5*time.Second, false)
	if err != nil {
		t.Fatalf("failed to create attacker: %v", err)
	}

	res := a.Attack(context.Background())

	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
}

func TestScriptAttacker_HeadersRequired_Missing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Script without headers — server should reject
	starScript := `
def request():
    return {
        "method": "GET",
        "url": "` + srv.URL + `",
    }
`
	path := writeStarFile(t, starScript)
	a, err := script.NewScriptAttacker(path, 5*time.Second, false)
	if err != nil {
		t.Fatalf("failed to create attacker: %v", err)
	}

	res := a.Attack(context.Background())

	if res.Code != 401 {
		t.Errorf("expected 401 without auth header, got %d", res.Code)
	}
}
