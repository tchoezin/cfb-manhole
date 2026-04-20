import sqlite3
from pathlib import Path

from data import make_default_league

DB_PATH = Path(__file__).parent / "cfb_manhole.db"


def get_conn() -> sqlite3.Connection:
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    conn.execute("PRAGMA foreign_keys = ON")
    return conn


def init_db() -> None:
    with get_conn() as conn:
        conn.executescript("""
            CREATE TABLE IF NOT EXISTS leagues (
                id   TEXT PRIMARY KEY,
                name TEXT NOT NULL
            );
            CREATE TABLE IF NOT EXISTS divisions (
                id        INTEGER PRIMARY KEY AUTOINCREMENT,
                league_id TEXT    NOT NULL,
                name      TEXT    NOT NULL,
                FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE,
                UNIQUE (league_id, name)
            );
            CREATE TABLE IF NOT EXISTS players (
                id          INTEGER PRIMARY KEY AUTOINCREMENT,
                name        TEXT    NOT NULL,
                division_id INTEGER NOT NULL,
                FOREIGN KEY (division_id) REFERENCES divisions(id) ON DELETE CASCADE
            );
            CREATE TABLE IF NOT EXISTS player_teams (
                player_id INTEGER NOT NULL,
                team      TEXT    NOT NULL,
                FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
            );
            CREATE TABLE IF NOT EXISTS game_results (
                id        INTEGER PRIMARY KEY AUTOINCREMENT,
                league_id TEXT    NOT NULL,
                winner    TEXT    NOT NULL,
                loser     TEXT    NOT NULL,
                FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE
            );
            CREATE TABLE IF NOT EXISTS team_conferences (
                league_id  TEXT NOT NULL,
                team       TEXT NOT NULL,
                conference TEXT NOT NULL,
                PRIMARY KEY (league_id, team),
                FOREIGN KEY (league_id) REFERENCES leagues(id) ON DELETE CASCADE
            );
        """)


def seed_league(conn: sqlite3.Connection, league_dict: dict) -> None:
    league_id = league_dict["id"]
    conn.execute(
        "INSERT INTO leagues (id, name) VALUES (?, ?)",
        (league_id, league_dict["name"]),
    )

    for division_name, players in league_dict["divisions"].items():
        cur = conn.execute(
            "INSERT INTO divisions (league_id, name) VALUES (?, ?)",
            (league_id, division_name),
        )
        division_id = cur.lastrowid

        for player_name in players:
            cur = conn.execute(
                "INSERT INTO players (name, division_id) VALUES (?, ?)",
                (player_name, division_id),
            )
            player_id = cur.lastrowid

            picks = league_dict["player_picks"].get(division_name, {}).get(player_name, [])
            conn.executemany(
                "INSERT INTO player_teams (player_id, team) VALUES (?, ?)",
                [(player_id, team) for team in picks],
            )

    for winner, loser in league_dict["game_results"]:
        conn.execute(
            "INSERT INTO game_results (league_id, winner, loser) VALUES (?, ?, ?)",
            (league_id, winner, loser),
        )

    for team, conference in league_dict["team_conferences"].items():
        conn.execute(
            "INSERT INTO team_conferences (league_id, team, conference) VALUES (?, ?, ?)",
            (league_id, team, conference),
        )


# ── Leagues ──────────────────────────────────────────────────────────────────

def league_exists(conn: sqlite3.Connection, league_id: str) -> bool:
    return conn.execute(
        "SELECT 1 FROM leagues WHERE id = ?", (league_id,)
    ).fetchone() is not None


def get_league(conn: sqlite3.Connection, league_id: str):
    return conn.execute(
        "SELECT id, name FROM leagues WHERE id = ?", (league_id,)
    ).fetchone()


def list_leagues(conn: sqlite3.Connection):
    return conn.execute("SELECT id, name FROM leagues ORDER BY id").fetchall()


def create_league(conn: sqlite3.Connection, league_id: str, name: str) -> None:
    conn.execute("INSERT INTO leagues (id, name) VALUES (?, ?)", (league_id, name))


def update_league_name(conn: sqlite3.Connection, league_id: str, name: str) -> None:
    conn.execute("UPDATE leagues SET name = ? WHERE id = ?", (name, league_id))


def delete_league(conn: sqlite3.Connection, league_id: str) -> None:
    conn.execute("DELETE FROM leagues WHERE id = ?", (league_id,))


def get_player_count_for_league(conn: sqlite3.Connection, league_id: str) -> int:
    row = conn.execute(
        """
        SELECT COUNT(*) AS cnt
        FROM players p
        JOIN divisions d ON p.division_id = d.id
        WHERE d.league_id = ?
        """,
        (league_id,),
    ).fetchone()
    return row["cnt"]


# ── Divisions ─────────────────────────────────────────────────────────────────

def list_divisions(conn: sqlite3.Connection, league_id: str):
    return conn.execute(
        "SELECT id, name FROM divisions WHERE league_id = ? ORDER BY id",
        (league_id,),
    ).fetchall()


def get_division(conn: sqlite3.Connection, league_id: str, division_name: str):
    return conn.execute(
        "SELECT id, name FROM divisions WHERE league_id = ? AND name = ?",
        (league_id, division_name),
    ).fetchone()


def create_division(conn: sqlite3.Connection, league_id: str, name: str) -> int:
    cur = conn.execute(
        "INSERT INTO divisions (league_id, name) VALUES (?, ?)", (league_id, name)
    )
    return cur.lastrowid


def rename_division(conn: sqlite3.Connection, division_id: int, new_name: str) -> None:
    conn.execute("UPDATE divisions SET name = ? WHERE id = ?", (new_name, division_id))


def delete_division(conn: sqlite3.Connection, division_id: int) -> None:
    conn.execute("DELETE FROM divisions WHERE id = ?", (division_id,))


def get_player_count_for_division(conn: sqlite3.Connection, division_id: int) -> int:
    row = conn.execute(
        "SELECT COUNT(*) AS cnt FROM players WHERE division_id = ?", (division_id,)
    ).fetchone()
    return row["cnt"]


# ── Players ───────────────────────────────────────────────────────────────────

def find_player(conn: sqlite3.Connection, league_id: str, player_name: str):
    """Returns a row with player id, name, division_id, division_name."""
    return conn.execute(
        """
        SELECT p.id, p.name, p.division_id, d.name AS division_name
        FROM players p
        JOIN divisions d ON p.division_id = d.id
        WHERE d.league_id = ? AND p.name = ?
        """,
        (league_id, player_name),
    ).fetchone()


def list_players_in_division(conn: sqlite3.Connection, division_id: int):
    return conn.execute(
        "SELECT id, name FROM players WHERE division_id = ? ORDER BY id",
        (division_id,),
    ).fetchall()


def create_player(conn: sqlite3.Connection, division_id: int, name: str) -> int:
    cur = conn.execute(
        "INSERT INTO players (name, division_id) VALUES (?, ?)", (name, division_id)
    )
    return cur.lastrowid


def rename_player(conn: sqlite3.Connection, player_id: int, new_name: str) -> None:
    conn.execute("UPDATE players SET name = ? WHERE id = ?", (new_name, player_id))


def move_player(conn: sqlite3.Connection, player_id: int, new_division_id: int) -> None:
    conn.execute(
        "UPDATE players SET division_id = ? WHERE id = ?", (new_division_id, player_id)
    )


def delete_player(conn: sqlite3.Connection, player_id: int) -> None:
    conn.execute("DELETE FROM players WHERE id = ?", (player_id,))


# ── Player teams ──────────────────────────────────────────────────────────────

def get_player_teams(conn: sqlite3.Connection, player_id: int) -> list[str]:
    rows = conn.execute(
        "SELECT team FROM player_teams WHERE player_id = ?", (player_id,)
    ).fetchall()
    return [row["team"] for row in rows]


def set_player_teams(conn: sqlite3.Connection, player_id: int, teams: list[str]) -> None:
    conn.execute("DELETE FROM player_teams WHERE player_id = ?", (player_id,))
    conn.executemany(
        "INSERT INTO player_teams (player_id, team) VALUES (?, ?)",
        [(player_id, team) for team in teams],
    )


# ── Games ─────────────────────────────────────────────────────────────────────

def list_games(conn: sqlite3.Connection, league_id: str):
    return conn.execute(
        "SELECT id, winner, loser FROM game_results WHERE league_id = ? ORDER BY id",
        (league_id,),
    ).fetchall()


def add_game(conn: sqlite3.Connection, league_id: str, winner: str, loser: str) -> int:
    cur = conn.execute(
        "INSERT INTO game_results (league_id, winner, loser) VALUES (?, ?, ?)",
        (league_id, winner, loser),
    )
    return cur.lastrowid


# ── Team conferences ──────────────────────────────────────────────────────────

def get_team_conferences(conn: sqlite3.Connection, league_id: str) -> dict[str, str]:
    rows = conn.execute(
        "SELECT team, conference FROM team_conferences WHERE league_id = ?",
        (league_id,),
    ).fetchall()
    return {row["team"]: row["conference"] for row in rows}


# ── Scoring helper ────────────────────────────────────────────────────────────

def load_league_dict(conn: sqlite3.Connection, league_id: str) -> dict:
    """Assemble the full league structure used by scoring functions."""
    league_row = get_league(conn, league_id)
    divisions_map: dict[str, list[str]] = {}
    player_picks_map: dict[str, dict[str, list[str]]] = {}

    for div_row in list_divisions(conn, league_id):
        div_name = div_row["name"]
        players_in_div: list[str] = []
        picks_in_div: dict[str, list[str]] = {}

        for p_row in list_players_in_division(conn, div_row["id"]):
            players_in_div.append(p_row["name"])
            picks_in_div[p_row["name"]] = get_player_teams(conn, p_row["id"])

        divisions_map[div_name] = players_in_div
        player_picks_map[div_name] = picks_in_div

    game_results = [
        (row["winner"], row["loser"]) for row in list_games(conn, league_id)
    ]

    return {
        "id": league_row["id"],
        "name": league_row["name"],
        "divisions": divisions_map,
        "player_picks": player_picks_map,
        "team_conferences": get_team_conferences(conn, league_id),
        "game_results": game_results,
    }
