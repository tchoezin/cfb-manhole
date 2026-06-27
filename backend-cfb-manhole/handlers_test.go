package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRequestAllowsConfiguredPreflightOrigin(t *testing.T) {
	t.Parallel()

	corsPolicy, err := newCORSPolicy([]string{"http://localhost:3000"})
	if err != nil {
		t.Fatalf("newCORSPolicy returned error: %v", err)
	}

	handler := handleRequest(NewRankingService(fakeRankingRepo{}), NewPlayerService(&fakePlayerRepo{}), corsPolicy)
	request := httptest.NewRequest(http.MethodOptions, "/api/players", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "Content-Type")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected allow origin header, got %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Fatalf("unexpected allow methods header %q", got)
	}
}

func TestHandleRequestRejectsDisallowedPreflightOrigin(t *testing.T) {
	t.Parallel()

	corsPolicy, err := newCORSPolicy([]string{"http://localhost:3000"})
	if err != nil {
		t.Fatalf("newCORSPolicy returned error: %v", err)
	}

	handler := handleRequest(NewRankingService(fakeRankingRepo{}), NewPlayerService(&fakePlayerRepo{}), corsPolicy)
	request := httptest.NewRequest(http.MethodOptions, "/api/players", nil)
	request.Header.Set("Origin", "https://evil.example")
	request.Header.Set("Access-Control-Request-Method", http.MethodDelete)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow origin header, got %q", got)
	}
	if vary := response.Header().Values("Vary"); len(vary) == 0 {
		t.Fatal("expected vary headers to be present")
	}
}

func TestLoadCORSPolicyDefaultsToLocalDevOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")

	corsPolicy, err := loadCORSPolicy()
	if err != nil {
		t.Fatalf("loadCORSPolicy returned error: %v", err)
	}

	if _, ok := corsPolicy.allowedOrigins["http://localhost:3000"]; !ok {
		t.Fatal("expected localhost dev origin to be allowed")
	}
	if _, ok := corsPolicy.allowedOrigins["http://127.0.0.1:3000"]; !ok {
		t.Fatal("expected 127.0.0.1 dev origin to be allowed")
	}
}

func TestLoadCORSPolicyRejectsInvalidOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "not-a-valid-origin")

	if _, err := loadCORSPolicy(); err == nil {
		t.Fatal("expected invalid origin error")
	}
}
