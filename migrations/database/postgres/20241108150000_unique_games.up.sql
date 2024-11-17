BEGIN;

ALTER TABLE games
    ADD COLUMN league_id INT REFERENCES leagues(id);

ALTER TABLE games
    ADD CONSTRAINT unique_teams_league UNIQUE (team1_id, team2_id, league_id);

UPDATE games g
SET league_id = tll.league_id
FROM teams_leagues_links tll
WHERE g.team1_id = tll.team_id;

COMMIT;