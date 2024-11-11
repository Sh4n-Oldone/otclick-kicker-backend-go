BEGIN;

ALTER TABLE games
    DROP CONSTRAINT unique_teams_league;

ALTER TABLE games
    DROP COLUMN league_id;

COMMIT;