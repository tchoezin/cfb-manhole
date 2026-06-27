package main

import (
	"context"
	"errors"
	"slices"
	"strings"
)

var ErrPlayerNotFound = errors.New("player not found")
var ErrPlayerAlreadyExists = errors.New("player already exists")

type rankingRepository interface {
	OrderedRankings(ctx context.Context) ([]playerEntry, error)
}

type playerRepository interface {
	ListPlayers(ctx context.Context) ([]playerRecord, error)
	GetPlayer(ctx context.Context, playerID string) (playerRecord, error)
	CreatePlayer(ctx context.Context, input createPlayerInput) (playerRecord, error)
	UpdatePlayerScore(ctx context.Context, playerID string, score int) (playerRecord, error)
	ReplacePlayerTeams(ctx context.Context, playerID string, teams []string) (playerRecord, error)
	DeletePlayer(ctx context.Context, playerID string) error
}

type createPlayerInput struct {
	Name     string
	Division int
	Score    int
	Teams    []string
}

type playerRecord struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Division int      `json:"division"`
	Score    int      `json:"score"`
	Teams    []string `json:"teams"`
}

type RankingService struct {
	repo rankingRepository
}

type PlayerService struct {
	repo playerRepository
}

func NewRankingService(repo rankingRepository) *RankingService {
	return &RankingService{repo: repo}
}

func NewPlayerService(repo playerRepository) *PlayerService {
	return &PlayerService{repo: repo}
}

func (s *RankingService) GetRankings(ctx context.Context) ([]playerEntry, error) {
	return s.repo.OrderedRankings(ctx)
}

func (s *RankingService) GetDivisions(ctx context.Context) (map[int][]playerEntry, error) {
	rankings, err := s.repo.OrderedRankings(ctx)
	if err != nil {
		return nil, err
	}

	return getDivisionMap(rankings), nil
}

func (s *RankingService) GetDivisionByID(ctx context.Context, id int) ([]playerEntry, error) {
	divisionMap, err := s.GetDivisions(ctx)
	if err != nil {
		return nil, err
	}

	return divisionMap[id], nil
}

func (s *PlayerService) ListPlayers(ctx context.Context) ([]playerRecord, error) {
	return s.repo.ListPlayers(ctx)
}

func (s *PlayerService) GetPlayer(ctx context.Context, playerID string) (playerRecord, error) {
	return s.repo.GetPlayer(ctx, strings.TrimSpace(playerID))
}

func (s *PlayerService) CreatePlayer(ctx context.Context, input createPlayerInput) (playerRecord, error) {
	normalized, err := normalizeCreatePlayerInput(input)
	if err != nil {
		return playerRecord{}, err
	}

	return s.repo.CreatePlayer(ctx, normalized)
}

func (s *PlayerService) UpdatePlayerScore(ctx context.Context, playerID string, score int) (playerRecord, error) {
	if strings.TrimSpace(playerID) == "" {
		return playerRecord{}, errors.New("player id is required")
	}
	if score < 0 {
		return playerRecord{}, errors.New("score must be zero or greater")
	}

	return s.repo.UpdatePlayerScore(ctx, strings.TrimSpace(playerID), score)
}

func (s *PlayerService) ReplacePlayerTeams(ctx context.Context, playerID string, teams []string) (playerRecord, error) {
	if strings.TrimSpace(playerID) == "" {
		return playerRecord{}, errors.New("player id is required")
	}

	normalizedTeams, err := normalizeTeams(teams)
	if err != nil {
		return playerRecord{}, err
	}

	return s.repo.ReplacePlayerTeams(ctx, strings.TrimSpace(playerID), normalizedTeams)
}

func (s *PlayerService) DeletePlayer(ctx context.Context, playerID string) error {
	if strings.TrimSpace(playerID) == "" {
		return errors.New("player id is required")
	}

	return s.repo.DeletePlayer(ctx, strings.TrimSpace(playerID))
}

func normalizeCreatePlayerInput(input createPlayerInput) (createPlayerInput, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return createPlayerInput{}, errors.New("name is required")
	}
	if input.Division <= 0 {
		return createPlayerInput{}, errors.New("division must be greater than zero")
	}
	if input.Score < 0 {
		return createPlayerInput{}, errors.New("score must be zero or greater")
	}

	teams, err := normalizeTeams(input.Teams)
	if err != nil {
		return createPlayerInput{}, err
	}

	return createPlayerInput{
		Name:     name,
		Division: input.Division,
		Score:    input.Score,
		Teams:    teams,
	}, nil
}

func normalizeTeams(teams []string) ([]string, error) {
	if len(teams) == 0 {
		return nil, errors.New("at least one team is required")
	}

	seen := make(map[string]struct{}, len(teams))
	normalized := make([]string, 0, len(teams))
	for _, team := range teams {
		trimmed := strings.TrimSpace(team)
		if trimmed == "" {
			return nil, errors.New("team names must not be empty")
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	slices.Sort(normalized)
	return normalized, nil
}

func getDivisionMap(playerRanking []playerEntry) map[int][]playerEntry {
	divisions := make(map[int][]playerEntry)

	for _, player := range playerRanking {
		divisions[player.Division] = append(divisions[player.Division], player)
	}

	divisionRankings := make(map[int][]playerEntry)
	for division, players := range divisions {
		divisionRankings[division] = rerank(players)
	}

	return divisionRankings
}

func rerank(entries []playerEntry) []playerEntry {
	output := make([]playerEntry, len(entries))
	for i, player := range entries {
		output[i] = playerEntry{
			Rank:     i + 1,
			Player:   player.Player,
			Score:    player.Score,
			Division: player.Division,
		}
	}
	return output
}
