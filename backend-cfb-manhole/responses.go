package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var defaultAllowedOrigins = []string{
	"http://127.0.0.1:3000",
	"http://localhost:3000",
}

type CORSPolicy struct {
	allowedOrigins map[string]struct{}
}

func loadCORSPolicy() (*CORSPolicy, error) {
	rawOrigins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if rawOrigins == "" {
		return newCORSPolicy(defaultAllowedOrigins)
	}

	parts := strings.Split(rawOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return nil, fmt.Errorf("CORS_ALLOWED_ORIGINS must contain at least one origin")
	}

	return newCORSPolicy(origins)
}

func newCORSPolicy(origins []string) (*CORSPolicy, error) {
	allowedOrigins := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		normalizedOrigin, err := normalizeOrigin(origin)
		if err != nil {
			return nil, err
		}
		allowedOrigins[normalizedOrigin] = struct{}{}
	}

	return &CORSPolicy{allowedOrigins: allowedOrigins}, nil
}

func normalizeOrigin(origin string) (string, error) {
	normalizedOrigin := strings.TrimSpace(strings.TrimRight(origin, "/"))
	if normalizedOrigin == "" {
		return "", fmt.Errorf("origin cannot be empty")
	}

	parsedOrigin, err := url.Parse(normalizedOrigin)
	if err != nil {
		return "", fmt.Errorf("invalid origin %q: %w", origin, err)
	}
	if parsedOrigin.Scheme == "" || parsedOrigin.Host == "" || parsedOrigin.Path != "" {
		return "", fmt.Errorf("invalid origin %q", origin)
	}

	return parsedOrigin.String(), nil
}

func (policy *CORSPolicy) apply(w http.ResponseWriter, r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}

	addVaryHeader(w, "Origin")
	if r.Method == http.MethodOptions {
		addVaryHeader(w, "Access-Control-Request-Method")
		addVaryHeader(w, "Access-Control-Request-Headers")
	}

	normalizedOrigin, err := normalizeOrigin(origin)
	if err != nil {
		return false
	}
	if _, ok := policy.allowedOrigins[normalizedOrigin]; !ok {
		return false
	}

	w.Header().Set("Access-Control-Allow-Origin", normalizedOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	return true
}

func addVaryHeader(w http.ResponseWriter, value string) {
	w.Header().Add("Vary", value)
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return io.ErrUnexpectedEOF
	}

	return nil
}

func sendJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
