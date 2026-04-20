from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

from data import all_teams, make_default_league
from db import (
    add_game,
    create_division,
    create_league,
    create_player,
    delete_division,
    delete_league,
    delete_player,
    find_player,
    get_conn,
    get_division,
    get_league,
    get_player_count_for_division,
    get_player_count_for_league,
    get_player_teams,
    init_db,
    league_exists,
    list_divisions,
    list_games,
    list_leagues,
    list_players_in_division,
    load_league_dict,
    move_player,
    rename_division,
    rename_player,
    seed_league,
    set_player_teams,
    update_league_name,
)
from scoring import build_division_leaderboards, calculate_scores


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    with get_conn() as conn:
        if not league_exists(conn, "default"):
            seed_league(conn, make_default_league())
    yield


app = FastAPI(title="CFB Manhole API", version="0.1.0", lifespan=lifespan)

# Open CORS keeps frontend integration simple for local/group usage.
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# ── Request models ────────────────────────────────────────────────────────────

class LeagueCreateRequest(BaseModel):
    league_id: str = Field(min_length=1)
    name: Optional[str] = None
    use_mock_data: bool = True


class LeagueUpdateRequest(BaseModel):
    name: str = Field(min_length=1)


class DivisionCreateRequest(BaseModel):
    division_name: str = Field(min_length=1)


class DivisionRenameRequest(BaseModel):
    new_name: str = Field(min_length=1)


class PlayerCreateRequest(BaseModel):
    player_name: str = Field(min_length=1)
    division_name: str = Field(min_length=1)
    teams: list[str] = Field(default_factory=list)


class PlayerUpdateRequest(BaseModel):
    new_name: Optional[str] = None
    division_name: Optional[str] = None
    teams: Optional[list[str]] = None


class TeamUpdateRequest(BaseModel):
    teams: list[str]


class GameResultCreateRequest(BaseModel):
    winner: str = Field(min_length=1)
    loser: str = Field(min_length=1)


# ── Shared helpers ────────────────────────────────────────────────────────────

def _require_league(conn, league_id: str):
    row = get_league(conn, league_id)
    if not row:
        raise HTTPException(status_code=404, detail=f"League '{league_id}' not found")
    return row


def _require_division(conn, league_id: str, division_name: str):
    row = get_division(conn, league_id, division_name)
    if not row:
        raise HTTPException(status_code=404, detail="Division not found")
    return row


def _require_player(conn, league_id: str, player_name: str):
    row = find_player(conn, league_id, player_name)
    if not row:
        raise HTTPException(status_code=404, detail="Player not found")
    return row


def _validate_team_names(teams: list[str]):
    unknown = [team for team in teams if team not in all_teams]
    if unknown:
        raise HTTPException(
            status_code=400,
            detail={"message": "Unknown teams provided", "unknown_teams": unknown},
        )


def _compute_scores(league_dict: dict) -> dict:
    return calculate_scores(
        divisions=league_dict["divisions"],
        player_picks=league_dict["player_picks"],
        team_conferences=league_dict["team_conferences"],
        game_results=league_dict["game_results"],
    )


# ── Utility endpoints ─────────────────────────────────────────────────────────

@app.get("/health")
def health_check():
    return {"status": "ok"}


@app.get("/")
def root():
    return {"service": "cfb-manhole-backend", "version": "0.1.0", "docs": "/docs"}


# ── Leagues ───────────────────────────────────────────────────────────────────

@app.get("/leagues")
def list_all_leagues():
    with get_conn() as conn:
        rows = list_leagues(conn)
        return [
            {
                "id": row["id"],
                "name": row["name"],
                "player_count": get_player_count_for_league(conn, row["id"]),
                "divisions": [d["name"] for d in list_divisions(conn, row["id"])],
            }
            for row in rows
        ]


@app.post("/leagues", status_code=201)
def create_new_league(payload: LeagueCreateRequest):
    with get_conn() as conn:
        if league_exists(conn, payload.league_id):
            raise HTTPException(status_code=409, detail="League already exists")

        if payload.use_mock_data:
            league_dict = make_default_league(
                league_id=payload.league_id,
                league_name=payload.name or payload.league_id,
            )
            seed_league(conn, league_dict)
        else:
            create_league(conn, payload.league_id, payload.name or payload.league_id)

        return load_league_dict(conn, payload.league_id)


@app.get("/leagues/{league_id}")
def get_one_league(league_id: str):
    with get_conn() as conn:
        _require_league(conn, league_id)
        return load_league_dict(conn, league_id)


@app.put("/leagues/{league_id}")
def update_league(league_id: str, payload: LeagueUpdateRequest):
    with get_conn() as conn:
        _require_league(conn, league_id)
        update_league_name(conn, league_id, payload.name)
        return {"id": league_id, "name": payload.name}


@app.delete("/leagues/{league_id}")
def delete_one_league(league_id: str, force: bool = Query(default=False)):
    with get_conn() as conn:
        league_row = _require_league(conn, league_id)
        player_count = get_player_count_for_league(conn, league_id)

        if player_count > 0 and not force:
            raise HTTPException(
                status_code=400,
                detail={
                    "message": "League has players. Delete players first or set force=true.",
                    "league_id": league_id,
                    "player_count": player_count,
                },
            )

        delete_league(conn, league_id)
        return {
            "message": "League deleted",
            "league_id": league_id,
            "league_name": league_row["name"],
        }


# ── Divisions ─────────────────────────────────────────────────────────────────

@app.get("/leagues/{league_id}/divisions")
def list_all_divisions(league_id: str):
    with get_conn() as conn:
        _require_league(conn, league_id)
        return {row["name"]: [] for row in list_divisions(conn, league_id)}


@app.post("/leagues/{league_id}/divisions", status_code=201)
def create_new_division(league_id: str, payload: DivisionCreateRequest):
    with get_conn() as conn:
        _require_league(conn, league_id)
        if get_division(conn, league_id, payload.division_name):
            raise HTTPException(status_code=409, detail="Division already exists")

        create_division(conn, league_id, payload.division_name)
        return {"division_name": payload.division_name}


@app.put("/leagues/{league_id}/divisions/{division_name}")
def rename_one_division(league_id: str, division_name: str, payload: DivisionRenameRequest):
    with get_conn() as conn:
        _require_league(conn, league_id)
        div_row = _require_division(conn, league_id, division_name)

        if get_division(conn, league_id, payload.new_name):
            raise HTTPException(status_code=409, detail="New division name already exists")

        rename_division(conn, div_row["id"], payload.new_name)
        return {"old_name": division_name, "new_name": payload.new_name}


@app.delete("/leagues/{league_id}/divisions/{division_name}")
def delete_one_division(
    league_id: str,
    division_name: str,
    force: bool = Query(default=False),
):
    with get_conn() as conn:
        _require_league(conn, league_id)
        div_row = _require_division(conn, league_id, division_name)
        player_count = get_player_count_for_division(conn, div_row["id"])

        if player_count > 0 and not force:
            raise HTTPException(
                status_code=400,
                detail={
                    "message": "Division has players. Set force=true to delete it.",
                    "division_name": division_name,
                    "player_count": player_count,
                },
            )

        removed_players = [
            p["name"] for p in list_players_in_division(conn, div_row["id"])
        ]
        delete_division(conn, div_row["id"])
        return {
            "message": "Division deleted",
            "division_name": division_name,
            "removed_players": removed_players,
        }


# ── Players ───────────────────────────────────────────────────────────────────

@app.get("/leagues/{league_id}/players")
def list_all_players(league_id: str):
    with get_conn() as conn:
        _require_league(conn, league_id)
        league_dict = load_league_dict(conn, league_id)
        scores = _compute_scores(league_dict)

        players = []
        for division_name, picks in league_dict["player_picks"].items():
            for player_name, teams in picks.items():
                players.append(
                    {
                        "player_name": player_name,
                        "division_name": division_name,
                        "teams": teams,
                        "score": scores.get(player_name, 0),
                    }
                )
        return sorted(players, key=lambda r: r["score"], reverse=True)


@app.post("/leagues/{league_id}/players", status_code=201)
def create_new_player(league_id: str, payload: PlayerCreateRequest):
    with get_conn() as conn:
        _require_league(conn, league_id)
        div_row = _require_division(conn, league_id, payload.division_name)

        if find_player(conn, league_id, payload.player_name):
            raise HTTPException(status_code=409, detail="Player already exists")

        _validate_team_names(payload.teams)
        player_id = create_player(conn, div_row["id"], payload.player_name)
        set_player_teams(conn, player_id, payload.teams)

        return {
            "player_name": payload.player_name,
            "division_name": payload.division_name,
            "teams": payload.teams,
        }


@app.put("/leagues/{league_id}/players/{player_name}")
def update_one_player(league_id: str, player_name: str, payload: PlayerUpdateRequest):
    with get_conn() as conn:
        _require_league(conn, league_id)
        player_row = _require_player(conn, league_id, player_name)

        target_name = payload.new_name or player_name
        target_division_name = payload.division_name or player_row["division_name"]

        target_div_row = get_division(conn, league_id, target_division_name)
        if not target_div_row:
            raise HTTPException(status_code=404, detail="Target division not found")

        if target_name != player_name and find_player(conn, league_id, target_name):
            raise HTTPException(status_code=409, detail="Target player name already exists")

        existing_teams = get_player_teams(conn, player_row["id"])
        updated_teams = payload.teams if payload.teams is not None else existing_teams
        _validate_team_names(updated_teams)

        if target_name != player_name:
            rename_player(conn, player_row["id"], target_name)
        if target_div_row["id"] != player_row["division_id"]:
            move_player(conn, player_row["id"], target_div_row["id"])
        set_player_teams(conn, player_row["id"], updated_teams)

        return {
            "player_name": target_name,
            "division_name": target_division_name,
            "teams": updated_teams,
        }


@app.delete("/leagues/{league_id}/players/{player_name}")
def delete_one_player(league_id: str, player_name: str):
    with get_conn() as conn:
        _require_league(conn, league_id)
        player_row = _require_player(conn, league_id, player_name)
        delete_player(conn, player_row["id"])
        return {
            "message": "Player deleted",
            "player_name": player_name,
            "division_name": player_row["division_name"],
        }


@app.put("/leagues/{league_id}/players/{player_name}/teams")
def update_player_teams(league_id: str, player_name: str, payload: TeamUpdateRequest):
    with get_conn() as conn:
        _require_league(conn, league_id)
        player_row = _require_player(conn, league_id, player_name)
        _validate_team_names(payload.teams)
        set_player_teams(conn, player_row["id"], payload.teams)
        return {
            "player_name": player_name,
            "division_name": player_row["division_name"],
            "teams": payload.teams,
        }


# ── Games ─────────────────────────────────────────────────────────────────────

@app.get("/leagues/{league_id}/games")
def list_all_games(league_id: str):
    with get_conn() as conn:
        _require_league(conn, league_id)
        return [
            {"game_id": row["id"], "winner": row["winner"], "loser": row["loser"]}
            for row in list_games(conn, league_id)
        ]


@app.post("/leagues/{league_id}/games", status_code=201)
def create_game_result(league_id: str, payload: GameResultCreateRequest):
    with get_conn() as conn:
        _require_league(conn, league_id)
        game_id = add_game(conn, league_id, payload.winner, payload.loser)
        return {"game_id": game_id, "winner": payload.winner, "loser": payload.loser}


# ── Scores & leaderboard ──────────────────────────────────────────────────────

@app.get("/leagues/{league_id}/scores")
def read_scores(league_id: str):
    with get_conn() as conn:
        _require_league(conn, league_id)
        league_dict = load_league_dict(conn, league_id)
        return {"league_id": league_id, "scores": _compute_scores(league_dict)}


@app.get("/leagues/{league_id}/leaderboard")
def read_leaderboard(league_id: str):
    with get_conn() as conn:
        _require_league(conn, league_id)
        league_dict = load_league_dict(conn, league_id)
        scores = _compute_scores(league_dict)
        return {
            "league_id": league_id,
            "leaderboard": build_division_leaderboards(league_dict["divisions"], scores),
        }


@app.post("/leagues/{league_id}/reset")
def reset_league_to_mock_data(league_id: str):
    with get_conn() as conn:
        league_row = _require_league(conn, league_id)
        delete_league(conn, league_id)
        seed_league(
            conn,
            make_default_league(
                league_id=league_id,
                league_name=league_row["name"],
            ),
        )
    return {"message": f"League '{league_id}' reset to mock data"}


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("backend:app", host="0.0.0.0", port=8000, reload=True)
