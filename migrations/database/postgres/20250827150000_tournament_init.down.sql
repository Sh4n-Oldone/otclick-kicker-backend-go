BEGIN;

DROP TABLE IF EXISTS tournaments_rating;

ALTER TABLE games DROP COLUMN stage_id;

DROP TABLE IF EXISTS tournaments_teams_link;
DROP TABLE IF EXISTS tournament_stages;
DROP TABLE IF EXISTS tournaments;
DROP TABLE IF EXISTS tournament_types;

COMMIT;