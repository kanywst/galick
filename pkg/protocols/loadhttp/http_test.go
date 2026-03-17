package loadhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kanywst/galick/pkg/protocols/loadhttp"
)

func TestAttacker_NoHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := loadhttp.NewAttacker("GET", srv.URL, 5*time.Second, false, nil, 2)
	res := a.Attack(context.Background())

	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
	if res.Error != "" {
		t.Errorf("unexpected error: %s", res.Error)
	}
}

func TestAttacker_WithHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	headers := map[string]string{
		"Authorization": "Bearer test-token",
		"Content-Type":  "application/json",
	}
	a := loadhttp.NewAttacker("GET", srv.URL, 5*time.Second, false, headers, 2)
	res := a.Attack(context.Background())

	if res.Code != 200 {
		t.Errorf("expected 200, got %d (headers not applied)", res.Code)
	}
	if res.Error != "" {
		t.Errorf("unexpected error: %s", res.Error)
	}
}

func TestAttacker_MissingHeaders_Rejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// No headers — server should reject
	a := loadhttp.NewAttacker("GET", srv.URL, 5*time.Second, false, nil, 2)
	res := a.Attack(context.Background())

	if res.Code != 401 {
		t.Errorf("expected 401 without auth header, got %d", res.Code)
	}
}
