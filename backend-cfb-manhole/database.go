package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDatabaseTimeout = 5 * time.Second
const seedDatabaseTimeout = 30 * time.Second

type postgresRepository struct {
	pool *pgxpool.Pool
}

func newPostgresRepository(ctx context.Context, databaseURL string) (*postgresRepository, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, defaultDatabaseTimeout)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &postgresRepository{pool: pool}, nil
}

func lookupDatabaseURL() (string, error) {
	if value := strings.TrimSpace(os.Getenv("NEON_DATABASE_URL")); value != "" {
		return value, nil
	}
	if value := strings.TrimSpace(os.Getenv("DATABASE_URL")); value != "" {
		return value, nil
	}

	return "", errors.New("NEON_DATABASE_URL or DATABASE_URL must be set")
}

func (r *postgresRepository) Close() {
	r.pool.Close()
}

func (r *postgresRepository) EnsureSchema(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultDatabaseTimeout)
	defer cancel()

	_, err := r.pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS players (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	division INTEGER NOT NULL,
	current_score INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS player_teams (
	player_id TEXT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
	team_name TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (player_id, team_name)
);

CREATE INDEX IF NOT EXISTS idx_players_division_score ON players (division, current_score DESC, name ASC);
`)
	if err != nil {
		return fmt.Errorf("ensure schema: %w", err)
	}

	return nil
}

func (r *postgresRepository) SeedInitialData(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, seedDatabaseTimeout)
	defer cancel()

	var existingCount int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM players`).Scan(&existingCount); err != nil {
		return fmt.Errorf("count players: %w", err)
	}
	if existingCount > 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, seed := range buildSeedPlayers() {
		playerID := uuid.NewString()
		batch.Queue(`
INSERT INTO players (id, name, division, current_score)
VALUES ($1, $2, $3, $4)
`, playerID, seed.Name, seed.Division, seed.Score)

		for _, team := range seed.Teams {
			batch.Queue(`
INSERT INTO player_teams (player_id, team_name)
VALUES ($1, $2)
`, playerID, team)
		}
	}

	results := tx.SendBatch(ctx, batch)
	if err := results.Close(); err != nil {
		return fmt.Errorf("execute seed batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	return nil
}

func (r *postgresRepository) OrderedRankings(ctx context.Context) ([]playerEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDatabaseTimeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
SELECT name, current_score, division
FROM players
ORDER BY current_score DESC, name ASC
`)
	if err != nil {
		return nil, fmt.Errorf("query rankings: %w", err)
	}
	defer rows.Close()

	entries := make([]playerEntry, 0)
	for rows.Next() {
		var entry playerEntry
		if err := rows.Scan(&entry.Player, &entry.Score, &entry.Division); err != nil {
			return nil, fmt.Errorf("scan ranking row: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ranking rows: %w", err)
	}

	for index := range entries {
		entries[index].Rank = index + 1
	}

	return entries, nil
}

func (r *postgresRepository) ListPlayers(ctx context.Context) ([]playerRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDatabaseTimeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
SELECT p.id, p.name, p.division, p.current_score,
	COALESCE(array_agg(pt.team_name ORDER BY pt.team_name) FILTER (WHERE pt.team_name IS NOT NULL), '{}') AS teams
FROM players p
LEFT JOIN player_teams pt ON pt.player_id = p.id
GROUP BY p.id, p.name, p.division, p.current_score
ORDER BY p.current_score DESC, p.name ASC
`)
	if err != nil {
		return nil, fmt.Errorf("query players: %w", err)
	}
	defer rows.Close()

	players := make([]playerRecord, 0)
	for rows.Next() {
		var player playerRecord
		if err := rows.Scan(&player.ID, &player.Name, &player.Division, &player.Score, &player.Teams); err != nil {
			return nil, fmt.Errorf("scan player row: %w", err)
		}
		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate player rows: %w", err)
	}

	return players, nil
}

func (r *postgresRepository) GetPlayer(ctx context.Context, playerID string) (playerRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDatabaseTimeout)
	defer cancel()

	player, err := fetchPlayer(ctx, r.pool, playerID)
	if err != nil {
		return playerRecord{}, err
	}

	return player, nil
}

func (r *postgresRepository) CreatePlayer(ctx context.Context, input createPlayerInput) (playerRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDatabaseTimeout)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return playerRecord{}, fmt.Errorf("begin create transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	playerID := uuid.NewString()
	_, err = tx.Exec(ctx, `
INSERT INTO players (id, name, division, current_score)
VALUES ($1, $2, $3, $4)
`, playerID, input.Name, input.Division, input.Score)
	if err != nil {
		if isUniqueViolation(err) {
			return playerRecord{}, ErrPlayerAlreadyExists
		}
		return playerRecord{}, fmt.Errorf("insert player: %w", err)
	}

	for _, team := range input.Teams {
		_, err = tx.Exec(ctx, `
INSERT INTO player_teams (player_id, team_name)
VALUES ($1, $2)
`, playerID, team)
		if err != nil {
			return playerRecord{}, fmt.Errorf("insert player team: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return playerRecord{}, fmt.Errorf("commit create transaction: %w", err)
	}

	return r.GetPlayer(ctx, playerID)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *postgresRepository) UpdatePlayerScore(ctx context.Context, playerID string, score int) (playerRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDatabaseTimeout)
	defer cancel()

	commandTag, err := r.pool.Exec(ctx, `
UPDATE players
SET current_score = $2, updated_at = NOW()
WHERE id = $1
`, playerID, score)
	if err != nil {
		return playerRecord{}, fmt.Errorf("update player score: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return playerRecord{}, ErrPlayerNotFound
	}

	return r.GetPlayer(ctx, playerID)
}

func (r *postgresRepository) ReplacePlayerTeams(ctx context.Context, playerID string, teams []string) (playerRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDatabaseTimeout)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return playerRecord{}, fmt.Errorf("begin team update transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	commandTag, err := tx.Exec(ctx, `
UPDATE players
SET updated_at = NOW()
WHERE id = $1
`, playerID)
	if err != nil {
		return playerRecord{}, fmt.Errorf("touch player during team update: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return playerRecord{}, ErrPlayerNotFound
	}

	if _, err := tx.Exec(ctx, `DELETE FROM player_teams WHERE player_id = $1`, playerID); err != nil {
		return playerRecord{}, fmt.Errorf("delete existing teams: %w", err)
	}

	for _, team := range teams {
		_, err = tx.Exec(ctx, `
INSERT INTO player_teams (player_id, team_name)
VALUES ($1, $2)
`, playerID, team)
		if err != nil {
			return playerRecord{}, fmt.Errorf("insert replacement team: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return playerRecord{}, fmt.Errorf("commit team update transaction: %w", err)
	}

	return r.GetPlayer(ctx, playerID)
}

func (r *postgresRepository) DeletePlayer(ctx context.Context, playerID string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultDatabaseTimeout)
	defer cancel()

	commandTag, err := r.pool.Exec(ctx, `DELETE FROM players WHERE id = $1`, playerID)
	if err != nil {
		return fmt.Errorf("delete player: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrPlayerNotFound
	}

	return nil
}

type playerQueryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func fetchPlayer(ctx context.Context, queryer playerQueryer, playerID string) (playerRecord, error) {
	var player playerRecord
	err := queryer.QueryRow(ctx, `
SELECT p.id, p.name, p.division, p.current_score,
	COALESCE(array_agg(pt.team_name ORDER BY pt.team_name) FILTER (WHERE pt.team_name IS NOT NULL), '{}') AS teams
FROM players p
LEFT JOIN player_teams pt ON pt.player_id = p.id
WHERE p.id = $1
GROUP BY p.id, p.name, p.division, p.current_score
`, playerID).Scan(&player.ID, &player.Name, &player.Division, &player.Score, &player.Teams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return playerRecord{}, ErrPlayerNotFound
		}
		return playerRecord{}, fmt.Errorf("get player: %w", err)
	}

	return player, nil
}

type seedPlayer struct {
	Name     string
	Division int
	Score    int
	Teams    []string
}

func buildSeedPlayers() []seedPlayer {
	scores := calculateScores()
	seeds := make([]seedPlayer, 0)
	for divisionID, players := range divisions {
		sortedPlayers := append([]string(nil), players...)
		sort.Strings(sortedPlayers)
		for _, player := range sortedPlayers {
			teams := append([]string(nil), playerPicks[divisionID][player]...)
			sort.Strings(teams)
			seeds = append(seeds, seedPlayer{
				Name:     player,
				Division: divisionID,
				Score:    scores[player],
				Teams:    teams,
			})
		}
	}

	sort.Slice(seeds, func(i, j int) bool {
		if seeds[i].Division != seeds[j].Division {
			return seeds[i].Division < seeds[j].Division
		}
		return seeds[i].Name < seeds[j].Name
	})

	return seeds
}
