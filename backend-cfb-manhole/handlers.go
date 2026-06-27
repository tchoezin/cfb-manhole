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

type createPlayerRequest struct {
	Name     string   `json:"name"`
	Division int      `json:"division"`
	Score    int      `json:"score"`
	Teams    []string `json:"teams"`
}

type updatePlayerScoreRequest struct {
	Score int `json:"score"`
}

type replacePlayerTeamsRequest struct {
	Teams []string `json:"teams"`
}

func handleRequest(rankingService *RankingService, playerService *PlayerService, corsPolicy *CORSPolicy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if corsPolicy != nil {
			originAllowed := corsPolicy.apply(w, r)
			if r.Method == http.MethodOptions && !originAllowed {
				http.Error(w, "CORS origin not allowed", http.StatusForbidden)
				return
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		path := r.URL.Path

		switch {
		case path == "/api/rankings":
			if r.Method != http.MethodGet {
				sendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
				return
			}

			rankings, err := rankingService.GetRankings(r.Context())
			if err != nil {
				sendJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load rankings"})
				return
			}

			sendJSON(w, http.StatusOK, map[string]any{
				"count":    len(rankings),
				"rankings": rankings,
			})
			return

		case path == "/api/rankings/divisions":
			if r.Method != http.MethodGet {
				sendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
				return
			}

			divisionRankings, err := rankingService.GetDivisions(r.Context())
			if err != nil {
				sendJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load divisions"})
				return
			}

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
			if r.Method != http.MethodGet {
				sendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
				return
			}

			singleDivisionRankings, err := rankingService.GetDivisionByID(r.Context(), divisionID)
			if err != nil {
				sendJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load division rankings"})
				return
			}
			if len(singleDivisionRankings) == 0 {
				sendJSON(w, http.StatusNotFound, map[string]string{"error": ErrDivisionNotFound.Error()})
				return
			}

			sendJSON(w, http.StatusOK, map[string]any{
				"division": divisionID,
				"count":    len(singleDivisionRankings),
				"rankings": singleDivisionRankings,
			})
			return

		case path == "/api/players":
			switch r.Method {
			case http.MethodGet:
				players, err := playerService.ListPlayers(r.Context())
				if err != nil {
					sendJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load players"})
					return
				}

				sendJSON(w, http.StatusOK, map[string]any{
					"count":   len(players),
					"players": players,
				})
				return

			case http.MethodPost:
				var payload createPlayerRequest
				if err := decodeJSON(r, &payload); err != nil {
					sendJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
					return
				}

				player, err := playerService.CreatePlayer(r.Context(), createPlayerInput{
					Name:     payload.Name,
					Division: payload.Division,
					Score:    payload.Score,
					Teams:    payload.Teams,
				})
				if err != nil {
					handlePlayerServiceError(w, err)
					return
				}

				sendJSON(w, http.StatusCreated, map[string]any{"player": player})
				return

			default:
				sendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
				return
			}

		case strings.HasPrefix(path, "/api/players/"):
			handlePlayerDetailRoute(w, r, path, playerService)
			return

		case path == "/health":
			if r.Method != http.MethodGet {
				sendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
				return
			}

			sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return

		default:
			sendJSON(w, http.StatusNotFound, map[string]string{"error": "Not Found"})
			return
		}
	}
}

func handlePlayerDetailRoute(w http.ResponseWriter, r *http.Request, path string, playerService *PlayerService) {
	segments := strings.Split(strings.TrimPrefix(path, "/api/players/"), "/")
	if len(segments) == 0 || segments[0] == "" {
		sendJSON(w, http.StatusNotFound, map[string]string{"error": "Not Found"})
		return
	}

	playerID := segments[0]
	if len(segments) == 1 {
		switch r.Method {
		case http.MethodGet:
			player, err := playerService.GetPlayer(r.Context(), playerID)
			if err != nil {
				handlePlayerServiceError(w, err)
				return
			}

			sendJSON(w, http.StatusOK, map[string]any{"player": player})
			return

		case http.MethodDelete:
			if err := playerService.DeletePlayer(r.Context(), playerID); err != nil {
				handlePlayerServiceError(w, err)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return

		default:
			sendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
			return
		}
	}

	if len(segments) != 2 {
		sendJSON(w, http.StatusNotFound, map[string]string{"error": "Not Found"})
		return
	}

	switch segments[1] {
	case "score":
		if r.Method != http.MethodPut {
			sendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
			return
		}

		var payload updatePlayerScoreRequest
		if err := decodeJSON(r, &payload); err != nil {
			sendJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		player, err := playerService.UpdatePlayerScore(r.Context(), playerID, payload.Score)
		if err != nil {
			handlePlayerServiceError(w, err)
			return
		}

		sendJSON(w, http.StatusOK, map[string]any{"player": player})
		return

	case "teams":
		if r.Method != http.MethodPut {
			sendJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
			return
		}

		var payload replacePlayerTeamsRequest
		if err := decodeJSON(r, &payload); err != nil {
			sendJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		player, err := playerService.ReplacePlayerTeams(r.Context(), playerID, payload.Teams)
		if err != nil {
			handlePlayerServiceError(w, err)
			return
		}

		sendJSON(w, http.StatusOK, map[string]any{"player": player})
		return

	default:
		sendJSON(w, http.StatusNotFound, map[string]string{"error": "Not Found"})
		return
	}
}

func handlePlayerServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrPlayerNotFound):
		sendJSON(w, http.StatusNotFound, map[string]string{"error": ErrPlayerNotFound.Error()})
	case errors.Is(err, ErrPlayerAlreadyExists):
		sendJSON(w, http.StatusConflict, map[string]string{"error": ErrPlayerAlreadyExists.Error()})
	default:
		if isValidationError(err) {
			sendJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		sendJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func isValidationError(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return strings.Contains(message, "required") ||
		strings.Contains(message, "greater than zero") ||
		strings.Contains(message, "zero or greater") ||
		strings.Contains(message, "must not be empty")
}
