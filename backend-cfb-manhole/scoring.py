from collections import defaultdict


def calculate_scores(divisions, player_picks, team_conferences, game_results):
    """Calculate cumulative scores for every player across game results."""
    scores = defaultdict(int)

    for winner, loser in game_results:
        winner_conf = team_conferences.get(winner)
        loser_conf = team_conferences.get(loser)

        for division_name in divisions:
            picks_in_division = player_picks.get(division_name, {})

            for player, picks in picks_in_division.items():
                if winner not in picks:
                    continue

                loser_picked_in_same_division = any(
                    loser in other_picks
                    for other_player, other_picks in picks_in_division.items()
                    if other_player != player
                )

                if loser_picked_in_same_division:
                    points = 3
                elif winner_conf and loser_conf and winner_conf == loser_conf:
                    points = 2
                else:
                    points = 1

                scores[player] += points

    return dict(scores)


def build_division_leaderboards(divisions, scores):
    """Return sorted leaderboards by division."""
    leaderboards = {}

    for division_name, players in divisions.items():
        ranked_players = sorted(
            players,
            key=lambda player: scores.get(player, 0),
            reverse=True,
        )
        leaderboards[division_name] = [
            {"player": player, "score": scores.get(player, 0)}
            for player in ranked_players
        ]

    return leaderboards
