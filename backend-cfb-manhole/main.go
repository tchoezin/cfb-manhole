package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type rankingEntry struct {
	Rank     int    `json:"rank"`
	Player   string `json:"player"`
	Score    int    `json:"score"`
	Division string `json:"division"`
}

func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func sendJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	setCommonHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func rerank(entries []rankingEntry) []rankingEntry {
	output := make([]rankingEntry, len(entries))
	for i, item := range entries {
		output[i] = rankingEntry{
			Rank:     i + 1,
			Player:   item.Player,
			Score:    item.Score,
			Division: item.Division,
		}
	}
	return output
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		setCommonHeaders(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
		return
	}

	rankings := getOrderedRankings()
	path := r.URL.Path

	switch {
	case path == "/api/rankings":
		sendJSON(w, http.StatusOK, map[string]any{
			"count":    len(rankings),
			"rankings": rankings,
		})
		return

	case path == "/api/rankings/divisions":
		divisionBuckets := make(map[string][]rankingEntry)
		for _, item := range rankings {
			divisionBuckets[item.Division] = append(divisionBuckets[item.Division], item)
		}

		divisionRankings := make(map[string][]rankingEntry)
		for division, entries := range divisionBuckets {
			divisionRankings[division] = rerank(entries)
		}

		sendJSON(w, http.StatusOK, map[string]any{
			"count":     len(divisionRankings),
			"divisions": divisionRankings,
		})
		return

	case strings.HasPrefix(path, "/api/rankings/division/"):
		rawDivision := strings.TrimPrefix(path, "/api/rankings/division/")
		divisionName, err := url.PathUnescape(rawDivision)
		if err != nil {
			sendJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid division name"})
			return
		}

		filtered := make([]rankingEntry, 0)
		for _, item := range rankings {
			if item.Division == divisionName {
				filtered = append(filtered, item)
			}
		}

		if len(filtered) == 0 {
			sendJSON(w, http.StatusNotFound, map[string]string{
				"error":    "Division not found",
				"division": divisionName,
			})
			return
		}

		sendJSON(w, http.StatusOK, map[string]any{
			"division": divisionName,
			"count":    len(filtered),
			"rankings": rerank(filtered),
		})
		return

	case path == "/health":
		sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return

	default:
		sendJSON(w, http.StatusNotFound, map[string]string{"error": "Not Found"})
		return
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)

	addr := "127.0.0.1:8000"
	log.Printf("Server running on http://%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
