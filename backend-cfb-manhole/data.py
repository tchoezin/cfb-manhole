# data.py

# Division names and player names
divisions = {
    'Division 1': ['Player A', 'Player B', 'Player C', 'Player D', 'Player E', 'Player F'],
    'Division 2': ['Player G', 'Player H', 'Player I', 'Player J', 'Player K', 'Player L'],
}

# Master pool of available teams
all_teams = [
    'Alabama', 'Georgia', 'LSU', 'Texas', 'Michigan', 'Oregon', 'USC', 'Notre Dame', 'Penn State', 'Clemson',
    'Ohio State', 'Florida State', 'Washington', 'Texas A&M', 'Iowa', 'Utah', 'Kansas State', 'Ole Miss', 'Tennessee', 'Oklahoma',
    'Mississippi State', 'Arkansas', 'Nebraska', 'Purdue', 'Miami', 'North Carolina', 'Louisville', 'UCF', 'Cincinnati', 'BYU',
    'South Carolina', 'Missouri', 'Kentucky', 'Minnesota', 'Maryland', 'West Virginia', 'Arizona State', 'Oregon State', 'Colorado', 'Stanford',
    'Syracuse', 'Virginia Tech', 'Wake Forest', 'Duke', 'Boston College', 'Houston', 'TCU', 'Baylor', 'Texas Tech', 'SMU',
    'Tulane', 'Memphis', 'Boise State', 'San Diego State', 'Fresno State', 'Air Force', 'Wyoming', 'App State', 'Marshall', 'Liberty'
]

# Example draft results — no overlap within a division, but slight differences across divisions
player_picks = {
    'Division 1': {
        'Player A': ['Alabama', 'Georgia', 'LSU', 'Texas', 'Michigan', 'Oregon', 'USC', 'Notre Dame', 'Penn State', 'Clemson'],
        'Player B': ['Ohio State', 'Florida State', 'Washington', 'Texas A&M', 'Iowa', 'Utah', 'Kansas State', 'Ole Miss', 'Tennessee', 'Oklahoma'],
        'Player C': ['Mississippi State', 'Arkansas', 'Nebraska', 'Purdue', 'Miami', 'North Carolina', 'Louisville', 'UCF', 'Cincinnati', 'BYU'],
        'Player D': ['South Carolina', 'Missouri', 'Kentucky', 'Minnesota', 'Maryland', 'West Virginia', 'Arizona State', 'Oregon State', 'Colorado', 'Stanford'],
        'Player E': ['Syracuse', 'Virginia Tech', 'Wake Forest', 'Duke', 'Boston College', 'Houston', 'TCU', 'Baylor', 'Texas Tech', 'SMU'],
        'Player F': ['Tulane', 'Memphis', 'Boise State', 'San Diego State', 'Fresno State', 'Air Force', 'Wyoming', 'App State', 'Marshall', 'Liberty'],
    },
    'Division 2': {
        'Player G': ['Alabama', 'Georgia', 'LSU', 'Texas', 'Michigan', 'Oregon', 'USC', 'Notre Dame', 'Penn State', 'Florida State'],  # swapped Clemson for Florida State
        'Player H': ['Ohio State', 'Clemson', 'Washington', 'Texas A&M', 'Iowa', 'Utah', 'Kansas State', 'Ole Miss', 'Tennessee', 'Oklahoma'],  # swapped FSU for Clemson
        'Player I': ['Mississippi State', 'Arkansas', 'Nebraska', 'Purdue', 'Miami', 'North Carolina', 'Louisville', 'UCF', 'Cincinnati', 'BYU'],  # same as Div 1 for testing duplicate detection
        'Player J': ['South Carolina', 'Missouri', 'Kentucky', 'Minnesota', 'Maryland', 'West Virginia', 'Arizona State', 'Oregon State', 'Colorado', 'Stanford'],  # same as Div 1
        'Player K': ['Syracuse', 'Virginia Tech', 'Wake Forest', 'Duke', 'Boston College', 'Houston', 'TCU', 'Baylor', 'Memphis', 'Liberty'],  # swapped Texas Tech and SMU for Memphis and Liberty
        'Player L': ['Tulane', 'Boise State', 'San Diego State', 'Fresno State', 'Air Force', 'Wyoming', 'App State', 'Marshall', 'Texas Tech', 'SMU'],  # swapped Memphis and Liberty for Texas Tech and SMU
    }
}

# Conferences and Divisions for scoring
team_conferences = {
    'Alabama': 'SEC', 'Georgia': 'SEC', 'LSU': 'SEC', 'Texas': 'Big 12', 'Michigan': 'Big Ten',
    'Oregon': 'Pac-12', 'USC': 'Pac-12', 'Notre Dame': 'Independent', 'Penn State': 'Big Ten', 'Clemson': 'ACC',
    'Florida State': 'ACC', 'Ohio State': 'Big Ten', 'Washington': 'Pac-12'
}
