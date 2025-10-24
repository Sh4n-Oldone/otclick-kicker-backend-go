BEGIN;

ALTER TABLE games ADD COLUMN is_tiebreak BOOLEAN default FALSE NOT NULL;

--удалить прежнее ограничение
ALTER TABLE games DROP CONSTRAINT IF EXISTS unique_teams_league;

--добавить уникальный индекс (на те игры, которые не tiebreak)
CREATE UNIQUE INDEX unique_teams_league_non_tiebreak
ON games (team1_id, team2_id, league_id)
WHERE is_tiebreak = false;

ALTER TABLE matches
ALTER COLUMN team1_id SET NOT NULL,
ALTER COLUMN team2_id SET NOT NULL,
ALTER COLUMN score_team1 SET NOT NULL,
ALTER COLUMN score_team2 SET NOT NULL;

COMMIT;