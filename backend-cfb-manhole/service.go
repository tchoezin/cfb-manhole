package main

import (
	"errors"
)

var ErrDivisionNotFound = errors.New("division not found")

type rankingRepository interface {
	OrderedRankings() []playerEntry
}

type RankingService struct {
	repo rankingRepository
}

func NewRankingService(repo rankingRepository) *RankingService {
	return &RankingService{repo: repo}
}

func (s *RankingService) GetDivisions() map[int][]playerEntry {
	rankings := s.repo.OrderedRankings()
	divisionMap := getDivisionMap(rankings)

	return divisionMap

}

func (s *RankingService) GetDivisionByID(id int) []playerEntry {
	rankings := s.repo.OrderedRankings()
	divisionMap := getDivisionMap(rankings)

	return divisionMap[id]
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

type scoringRepository struct{}

func (scoringRepository) OrderedRankings() []playerEntry {
	return getOrderedRankings()
}
