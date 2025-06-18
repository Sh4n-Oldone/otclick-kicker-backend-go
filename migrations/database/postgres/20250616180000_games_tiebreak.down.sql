BEGIN;

ALTER TABLE games DROP COLUMN is_tiebreak;

-- удалить уникальный индекс
DROP INDEX unique_teams_league_non_tiebreak;

--вернуть прежнее ограничение
ALTER TABLE games
    ADD CONSTRAINT unique_teams_league UNIQUE (team1_id, team2_id, league_id);

ALTER TABLE matches
ALTER COLUMN team1_id DROP NOT NULL,
ALTER COLUMN team2_id DROP NOT NULL,
ALTER COLUMN score_team1 DROP NOT NULL,
ALTER COLUMN score_team2 DROP NOT NULL;

COMMIT;