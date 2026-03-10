package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func seedCache() *Cache {
	c := NewCache()
	c.Set("will-luck/docker-sentinel", &PackageStats{
		Owner:         "Will-Luck",
		Package:       "docker-sentinel",
		TotalPulls:    433,
		LatestVersion: "2.11.1",
		Architectures: []string{"linux/amd64", "linux/arm64"},
		ScrapedAt:     1710000000,
	})
	return c
}

func TestBadgeHandler_Pulls(t *testing.T) {
	cache := seedCache()
	handler := newMux(cache)

	req := httptest.NewRequest("GET", "/will-luck/docker-sentinel/pulls.json", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var badge BadgeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &badge); err != nil {
		t.Fatal(err)
	}
	if badge.SchemaVersion != 1 {
		t.Errorf("schemaVersion = %d, want 1", badge.SchemaVersion)
	}
	if badge.Label != "ghcr pulls" {
		t.Errorf("label = %q, want %q", badge.Label, "ghcr pulls")
	}
	if badge.Message != "433" {
		t.Errorf("message = %q, want %q", badge.Message, "433")
	}
}

func TestBadgeHandler_Version(t *testing.T) {
	cache := seedCache()
	handler := newMux(cache)

	req := httptest.NewRequest("GET", "/will-luck/docker-sentinel/version.json", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var badge BadgeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &badge); err != nil {
		t.Fatal(err)
	}
	if badge.Message != "2.11.1" {
		t.Errorf("message = %q, want %q", badge.Message, "2.11.1")
	}
}

func TestBadgeHandler_Arch(t *testing.T) {
	cache := seedCache()
	handler := newMux(cache)

	req := httptest.NewRequest("GET", "/will-luck/docker-sentinel/arch.json", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	var badge BadgeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &badge); err != nil {
		t.Fatal(err)
	}
	if badge.Message != "amd64 | arm64" {
		t.Errorf("message = %q, want %q", badge.Message, "amd64 | arm64")
	}
}

func TestBadgeHandler_NotConfigured(t *testing.T) {
	cache := NewCache()
	handler := newMux(cache)

	req := httptest.NewRequest("GET", "/unknown/package/pulls.json", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestBadgeHandler_UnknownBadge(t *testing.T) {
	cache := seedCache()
	handler := newMux(cache)

	req := httptest.NewRequest("GET", "/will-luck/docker-sentinel/unknown.json", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}
