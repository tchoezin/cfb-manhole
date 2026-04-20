const DEFAULT_API_BASE_URL = "http://localhost:8000";

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? DEFAULT_API_BASE_URL;

export class ApiError extends Error {
  status: number;
  detail: unknown;

  constructor(message: string, status: number, detail: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.detail = detail;
  }
}

async function apiRequest<TResponse>(
  path: string,
  init?: RequestInit,
): Promise<TResponse> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  });

  const isJson = response.headers
    .get("content-type")
    ?.includes("application/json");

  const payload = isJson ? await response.json() : null;

  if (!response.ok) {
    const detail = isJson && payload ? payload : await response.text();
    throw new ApiError("Request failed", response.status, detail);
  }

  return payload as TResponse;
}

export type LeagueSummary = {
  id: string;
  name: string;
  player_count: number;
  divisions: string[];
};

export type League = {
  id: string;
  name: string;
  divisions: Record<string, string[]>;
  player_picks: Record<string, Record<string, string[]>>;
  team_conferences: Record<string, string>;
  game_results: Array<[string, string]>;
};

export type PlayerRow = {
  player_name: string;
  division_name: string;
  teams: string[];
  score: number;
};

export type ScoreResponse = {
  league_id: string;
  scores: Record<string, number>;
};

export type LeaderboardResponse = {
  league_id: string;
  leaderboard: Record<string, Array<{ player: string; score: number }>>;
};

export type LeagueCreatePayload = {
  league_id: string;
  name?: string;
  use_mock_data?: boolean;
};

export type LeagueUpdatePayload = {
  name: string;
};

export type DivisionCreatePayload = {
  division_name: string;
};

export type DivisionRenamePayload = {
  new_name: string;
};

export type PlayerCreatePayload = {
  player_name: string;
  division_name: string;
  teams?: string[];
};

export type PlayerUpdatePayload = {
  new_name?: string;
  division_name?: string;
  teams?: string[];
};

export type PlayerTeamsUpdatePayload = {
  teams: string[];
};

export const cfbApi = {
  listLeagues(): Promise<LeagueSummary[]> {
    return apiRequest<LeagueSummary[]>("/leagues");
  },

  createLeague(payload: LeagueCreatePayload): Promise<League> {
    return apiRequest<League>("/leagues", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  },

  getLeague(leagueId: string): Promise<League> {
    return apiRequest<League>(`/leagues/${encodeURIComponent(leagueId)}`);
  },

  updateLeague(leagueId: string, payload: LeagueUpdatePayload): Promise<{ id: string; name: string }> {
    return apiRequest<{ id: string; name: string }>(
      `/leagues/${encodeURIComponent(leagueId)}`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
    );
  },

  deleteLeague(leagueId: string, force = false): Promise<{ message: string; league_id: string; league_name: string }> {
    const forceQuery = force ? "?force=true" : "";
    return apiRequest<{ message: string; league_id: string; league_name: string }>(
      `/leagues/${encodeURIComponent(leagueId)}${forceQuery}`,
      { method: "DELETE" },
    );
  },

  listDivisions(leagueId: string): Promise<Record<string, string[]>> {
    return apiRequest<Record<string, string[]>>(
      `/leagues/${encodeURIComponent(leagueId)}/divisions`,
    );
  },

  createDivision(leagueId: string, payload: DivisionCreatePayload): Promise<{ division_name: string }> {
    return apiRequest<{ division_name: string }>(
      `/leagues/${encodeURIComponent(leagueId)}/divisions`,
      {
        method: "POST",
        body: JSON.stringify(payload),
      },
    );
  },

  renameDivision(
    leagueId: string,
    divisionName: string,
    payload: DivisionRenamePayload,
  ): Promise<{ old_name: string; new_name: string }> {
    return apiRequest<{ old_name: string; new_name: string }>(
      `/leagues/${encodeURIComponent(leagueId)}/divisions/${encodeURIComponent(divisionName)}`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
    );
  },

  deleteDivision(
    leagueId: string,
    divisionName: string,
    force = false,
  ): Promise<{ message: string; division_name: string; removed_players: string[] }> {
    const forceQuery = force ? "?force=true" : "";
    return apiRequest<{ message: string; division_name: string; removed_players: string[] }>(
      `/leagues/${encodeURIComponent(leagueId)}/divisions/${encodeURIComponent(divisionName)}${forceQuery}`,
      { method: "DELETE" },
    );
  },

  listPlayers(leagueId: string): Promise<PlayerRow[]> {
    return apiRequest<PlayerRow[]>(`/leagues/${encodeURIComponent(leagueId)}/players`);
  },

  createPlayer(leagueId: string, payload: PlayerCreatePayload): Promise<{ player_name: string; division_name: string; teams: string[] }> {
    return apiRequest<{ player_name: string; division_name: string; teams: string[] }>(
      `/leagues/${encodeURIComponent(leagueId)}/players`,
      {
        method: "POST",
        body: JSON.stringify(payload),
      },
    );
  },

  updatePlayer(
    leagueId: string,
    playerName: string,
    payload: PlayerUpdatePayload,
  ): Promise<{ player_name: string; division_name: string; teams: string[] }> {
    return apiRequest<{ player_name: string; division_name: string; teams: string[] }>(
      `/leagues/${encodeURIComponent(leagueId)}/players/${encodeURIComponent(playerName)}`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
    );
  },

  deletePlayer(
    leagueId: string,
    playerName: string,
  ): Promise<{ message: string; player_name: string; division_name: string }> {
    return apiRequest<{ message: string; player_name: string; division_name: string }>(
      `/leagues/${encodeURIComponent(leagueId)}/players/${encodeURIComponent(playerName)}`,
      { method: "DELETE" },
    );
  },

  updatePlayerTeams(
    leagueId: string,
    playerName: string,
    payload: PlayerTeamsUpdatePayload,
  ): Promise<{ player_name: string; division_name: string; teams: string[] }> {
    return apiRequest<{ player_name: string; division_name: string; teams: string[] }>(
      `/leagues/${encodeURIComponent(leagueId)}/players/${encodeURIComponent(playerName)}/teams`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
    );
  },

  getScores(leagueId: string): Promise<ScoreResponse> {
    return apiRequest<ScoreResponse>(`/leagues/${encodeURIComponent(leagueId)}/scores`);
  },

  getLeaderboard(leagueId: string): Promise<LeaderboardResponse> {
    return apiRequest<LeaderboardResponse>(
      `/leagues/${encodeURIComponent(leagueId)}/leaderboard`,
    );
  },
};
