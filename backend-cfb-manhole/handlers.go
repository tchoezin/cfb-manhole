package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

var ErrDivisionNotFound = errors.New("division not found")

type playerEntry struct {
	Rank     int    `json:"rank"`
	Player   string `json:"player"`
	Score    int    `json:"score"`
	Division int    `json:"division"`
}

func handleRequest(service *RankingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			divisionRankings := service.GetDivisions()

			sendJSON(w, http.StatusOK, map[string]any{
				"count":     len(divisionRankings),
				"divisions": divisionRankings,
			})
			return

		case strings.HasPrefix(path, "/api/rankings/divisions/"):
			idStr := strings.TrimPrefix(path, "/api/rankings/divisions/")
			divisionID, err := strconv.Atoi(idStr)
			if err != nil {
				sendJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid division id"})
				return
			}
			singleDivisionRankings := service.GetDivisionByID(divisionID)
			if len(singleDivisionRankings) == 0 {
				sendJSON(w, http.StatusNotFound, map[string]error{"error": ErrDivisionNotFound})
				return
			}

			sendJSON(w, http.StatusOK, map[string]any{
				"division": divisionID,
				"count":    len(singleDivisionRankings),
				"rankings": singleDivisionRankings,
			})
		case path == "/health":
			sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return

		default:
			sendJSON(w, http.StatusNotFound, map[string]string{"error": "Not Found"})
			return
		}
	}
}
