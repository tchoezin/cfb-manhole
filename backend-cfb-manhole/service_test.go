package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeRankingRepo struct {
	entries []playerEntry
	err     error
}

func (f fakeRankingRepo) OrderedRankings(context.Context) ([]playerEntry, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.entries, nil
}

type fakePlayerRepo struct {
	createdInput createPlayerInput
	createResult playerRecord
	createErr    error
	deletedID    string
	deleteErr    error
}

func (f *fakePlayerRepo) ListPlayers(context.Context) ([]playerRecord, error) {
	return nil, nil
}

func (f *fakePlayerRepo) GetPlayer(context.Context, string) (playerRecord, error) {
	return playerRecord{}, nil
}

func (f *fakePlayerRepo) CreatePlayer(_ context.Context, input createPlayerInput) (playerRecord, error) {
	f.createdInput = input
	if f.createErr != nil {
		return playerRecord{}, f.createErr
	}
	if f.createResult.ID == "" {
		f.createResult = playerRecord{
			ID:       "player-1",
			Name:     input.Name,
			Division: input.Division,
			Score:    input.Score,
			Teams:    input.Teams,
		}
	}
	return f.createResult, nil
}

func (f *fakePlayerRepo) UpdatePlayerScore(context.Context, string, int) (playerRecord, error) {
	return playerRecord{}, nil
}

func (f *fakePlayerRepo) ReplacePlayerTeams(context.Context, string, []string) (playerRecord, error) {
	return playerRecord{}, nil
}

func (f *fakePlayerRepo) DeletePlayer(_ context.Context, playerID string) error {
	f.deletedID = playerID
	return f.deleteErr
}

func TestRankingServiceGetDivisionsReranksWithinDivision(t *testing.T) {
	t.Parallel()

	service := NewRankingService(fakeRankingRepo{entries: []playerEntry{
		{Rank: 1, Player: "Taylor", Score: 12, Division: 1},
		{Rank: 2, Player: "Cam", Score: 11, Division: 2},
		{Rank: 3, Player: "Alcus", Score: 9, Division: 1},
	}})

	divisions, err := service.GetDivisions(context.Background())
	if err != nil {
		t.Fatalf("GetDivisions returned error: %v", err)
	}

	divisionOne := divisions[1]
	if len(divisionOne) != 2 {
		t.Fatalf("expected 2 players in division 1, got %d", len(divisionOne))
	}
	if divisionOne[0].Rank != 1 || divisionOne[0].Player != "Taylor" {
		t.Fatalf("expected Taylor reranked to 1, got %+v", divisionOne[0])
	}
	if divisionOne[1].Rank != 2 || divisionOne[1].Player != "Alcus" {
		t.Fatalf("expected Alcus reranked to 2, got %+v", divisionOne[1])
	}

	divisionTwo := divisions[2]
	if len(divisionTwo) != 1 || divisionTwo[0].Rank != 1 {
		t.Fatalf("expected division 2 to rerank from 1, got %+v", divisionTwo)
	}
}

func TestPlayerServiceCreatePlayerNormalizesInput(t *testing.T) {
	t.Parallel()

	repo := &fakePlayerRepo{}
	service := NewPlayerService(repo)

	player, err := service.CreatePlayer(context.Background(), createPlayerInput{
		Name:     "  Cho  ",
		Division: 1,
		Score:    4,
		Teams:    []string{"Texas", " Georgia ", "Texas"},
	})
	if err != nil {
		t.Fatalf("CreatePlayer returned error: %v", err)
	}

	if player.Name != "Cho" {
		t.Fatalf("expected trimmed player name, got %q", player.Name)
	}

	expectedTeams := []string{"Georgia", "Texas"}
	if !reflect.DeepEqual(repo.createdInput.Teams, expectedTeams) {
		t.Fatalf("expected normalized teams %v, got %v", expectedTeams, repo.createdInput.Teams)
	}
}

func TestPlayerServiceCreatePlayerRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	service := NewPlayerService(&fakePlayerRepo{})

	_, err := service.CreatePlayer(context.Background(), createPlayerInput{
		Name:     "",
		Division: 0,
		Score:    -1,
		Teams:    nil,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPlayerServicePropagatesDuplicatePlayerError(t *testing.T) {
	t.Parallel()

	repo := &fakePlayerRepo{createErr: ErrPlayerAlreadyExists}
	service := NewPlayerService(repo)

	_, err := service.CreatePlayer(context.Background(), createPlayerInput{
		Name:     "Taylor",
		Division: 1,
		Score:    3,
		Teams:    []string{"Georgia"},
	})
	if !errors.Is(err, ErrPlayerAlreadyExists) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestPlayerServiceDeletePlayerTrimsID(t *testing.T) {
	t.Parallel()

	repo := &fakePlayerRepo{}
	service := NewPlayerService(repo)

	if err := service.DeletePlayer(context.Background(), "  player-1  "); err != nil {
		t.Fatalf("DeletePlayer returned error: %v", err)
	}

	if repo.deletedID != "player-1" {
		t.Fatalf("expected trimmed player id, got %q", repo.deletedID)
	}
}
