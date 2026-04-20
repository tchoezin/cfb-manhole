from collections import defaultdict

def calculate_scores():
    team_conferences = {
        'Alabama': 'SEC', 'Georgia': 'SEC', 'Oregon': 'Pac-12', 'Texas': 'B12', 'Ohio State': 'B1G',
        'FSU': 'ACC', 'Notre Dame': 'Ind', 'Miami': 'ACC', 'Penn State': 'B1G', 'Tennessee': 'SEC',
        'Utah': 'Pac-12', 'Ole Miss': 'SEC', 'Michigan': 'B1G', 'Missouri': 'SEC', 'Kansas St': 'B12',
        'Clemson': 'ACC', 'LSU': 'SEC', 'NC State': 'ACC', 'Oklahoma': 'B12', 'Boise State': 'MWC',
        'Memphis': 'AAC', 'Texas a&m': 'SEC', 'Oklahoma St': 'B12', 'Liberty': 'CUSA', 'Iowa': 'B1G',
        'Kansas': 'B12', 'Louisville': 'ACC', 'James Madison': 'SB', 'Arizona': 'Pac-12', 'App St': 'SB',
        'Tulane': 'AAC', 'SMU': 'AAC', 'UCF': 'B12', 'UNLV': 'MWC', 'UTSA': 'AAC', 'Oregon St': 'Pac-12',
        'West Virginia': 'B12', 'Louisiana': 'SB', 'Miami (OH)': 'MAC', 'Toledo': 'MAC', 'Western Kentucky': 'CUSA',
        'Virginia Tech': 'ACC', 'Iowa St': 'B12', 'Texas State': 'SB', 'Nebraska': 'B1G', 'Wake Forest': 'ACC',
        'Bowling Green': 'MAC', 'Washington': 'Pac-12', 'Fresno State': 'MWC', 'Auburn': 'SEC', 'ECU': 'AAC',
        'USC': 'Pac-12', 'TCU': 'B12', 'Syracuse': 'ACC', 'UNC': 'ACC', 'Colorado': 'Pac-12', 'Wyoming': 'MWC',
        'Air Force': 'MWC', 'New Mexico State': 'CUSA', 'South Florida': 'AAC', 'Minnesota': 'B1G', 'Stanford': 'Pac-12',
        'Washington State': 'Pac-12', 'Coastal Carolina': 'SB', 'Wisconsin': 'B1G', 'Texas Tech': 'B12',
        'Colorado State': 'MWC'
    }

    divisions = {
        'Division 1': ['Taylor', 'Alcus', 'Ethan', 'JR', 'Cho', 'Gordie'],
        'Division 2': ['Cam', 'Teddy', 'Mike', 'Jason', 'Ian', 'Jack M'],
    }

    player_picks = {
        'Division 1': {
            'Taylor': ['Georgia', 'Notre Dame', 'Michigan', 'Oklahoma', 'Iowa', 'Tulane', 'Iowa St', 'Wisconsin', 'Colorado'],
            'Alcus': ['Oregon', 'Miami', 'Missouri', 'Boise State', 'Kansas', 'SMU', 'Louisiana', 'Texas State', 'Fresno State', 'Syracuse'],
            'Ethan': ['Texas', 'Penn State', 'Kansas St', 'Memphis', 'Louisville', 'UCF', 'Miami (OH)', 'Nebraska', 'Auburn', 'UNC'],
            'JR': ['Ohio State', 'Tennessee', 'Clemson', 'Texas a&m', 'James Madison', 'UNLV', 'Toledo', 'Wake Forest', 'ECU', 'TCU'],
            'Cho': ['Alabama', 'Utah', 'LSU', 'Oklahoma St', 'Arizona', 'UTSA', 'Western Kentucky', 'Bowling Green', 'USC', 'Wyoming'],
            'Gordie': ['FSU', 'Ole Miss', 'NC State', 'Liberty', 'App St', 'Oregon St', 'Virginia Tech', 'Washington', 'Texas Tech', 'Air Force'],
        },
        'Division 2': {
            'Cam': ['Georgia', 'LSU', 'Kansas St', 'Miami', 'Oklahoma', 'UCF', 'UNLV', 'Washington', 'UNC', 'Auburn'],
            'Teddy': ['Ohio State', 'Utah', 'Clemson', 'Tennessee', 'Tulane', 'Iowa', 'Iowa St', 'Western Kentucky', 'Texas State', 'Colorado State'],
            'Mike': ['Texas', 'FSU', 'Boise State', 'Kansas', 'App St', 'Texas a&m', 'Virginia Tech', 'UTSA', 'Louisiana', 'Washington State'],
            'Jason': ['Oregon', 'Michigan', 'Oklahoma St', 'SMU', 'Toledo', 'Louisville', 'James Madison', 'New Mexico State', 'Wyoming', 'Air Force'],
            'Ian': ['Ole Miss', 'Liberty', 'USC', 'Memphis', 'NC State', 'Miami (OH)', 'Nebraska', 'West Virginia', 'Wisconsin', 'South Florida'],
            'Jack M': ['Alabama', 'Penn State', 'Arizona', 'Missouri', 'Colorado', 'Notre Dame', 'Fresno State', 'Coastal Carolina', 'Minnesota', 'Stanford'],
        },
    }

    game_results = [
        ('NC State', 'Western Kentucky'), ('UCF', 'Colorado State'), ('Toledo', 'Bowling Green'),
        ('UNC', 'Michigan'), ('Coastal Carolina', 'App St'), ('Missouri', 'Louisiana'),
        ('Kansas', 'Texas State'), ('Colorado', 'Wyoming'), ('Tulane', 'Memphis'),
        ('Utah', 'Stanford'), ('Oklahoma', 'Temple'), ('Wisconsin', 'South Florida'),
        ('TCU', 'Stanford'), ('Georgia', 'Clemson'), ('Penn State', 'West Virginia'),
        ('Vanderbilt', 'Virginia Tech'), ('Iowa', 'Iowa St'), ('Louisville', 'James Madison'),
        ('Tennessee', 'Chattanooga'), ('Oklahoma St', 'Boise State')
    ]

    player_divisions = {}
    for div, players in divisions.items():
        for player in players:
            player_divisions[player] = div

    scores = defaultdict(int)

    for winner, loser in game_results:
        winner_conf = team_conferences.get(winner)
        loser_conf = team_conferences.get(loser)

        for div, picks_dict in player_picks.items():
            for player, picks in picks_dict.items():
                if winner in picks:
                    loser_picked_in_same_division = any(
                        loser in other_picks for other_player, other_picks in picks_dict.items() if other_player != player
                    )

                    if loser_picked_in_same_division:
                        points = 3
                    elif winner_conf and loser_conf and winner_conf == loser_conf:
                        points = 2
                    else:
                        points = 1

                    scores[player] += points

    for div in divisions:
        print(f"\n=== {div} Leaderboard ===")
        sorted_players = sorted(divisions[div], key=lambda p: scores[p], reverse=True)
        for player in sorted_players:
            print(f"{player}: {scores[player]} pts")

    return dict(scores)

calculate_scores()
