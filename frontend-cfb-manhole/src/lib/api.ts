export type RankingEntry = {
  rank: number;
  player: string;
  score: number;
  division: string;
};

export type RankingsResponse = {
  count: number;
  rankings: RankingEntry[];
};

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || "http://127.0.0.1:8000";

export async function FetchRankings() : Promise<RankingEntry[]> {
    const response = await fetch(API_BASE_URL + "/api/rankings", {
        method: "GET",
      });

    if (!response.ok) {
        throw new Error(
            "Request failed: " + response.status + " " + response.statusText
        );
    }

    const data: RankingsResponse = await response.json();

    return data.rankings
}