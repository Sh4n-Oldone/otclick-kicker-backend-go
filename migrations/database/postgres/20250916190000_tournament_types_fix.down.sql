BEGIN;

DELETE FROM tournament_types WHERE id IN (3,4,5);

DROP FUNCTION IF EXISTS get_next_name_suffix;

DROP SEQUENCE increment_bigint;

COMMIT;

/*
BEGIN;

-- Первый CTE для игр и турниров
WITH tournaments_ids AS (
    SELECT id FROM tournaments
    WHERE type_id IN (3,4,5)
),
stage_ids AS (
    SELECT id FROM tournament_stages
    WHERE tournament_id IN (SELECT id FROM tournaments_ids)
),
games_ids AS (
    SELECT id FROM games
    WHERE stage_id IN (SELECT id FROM stage_ids)
)
-- Удаление записей из matches
DELETE FROM matches
WHERE matches.game_id IN (SELECT id FROM games_ids);

-- Повторное использование CTE, т.к. область видимости каждого отдельного запроса изолирована
WITH tournaments_ids AS (
    SELECT id FROM tournaments
    WHERE type_id IN (3,4,5)
)
DELETE FROM tournament_rating
WHERE tournament_rating.tournament_id IN (SELECT id FROM tournaments_ids);

-- Повторное использование CTE для tournaments_teams_link
WITH tournaments_ids AS (
    SELECT id FROM tournaments
    WHERE type_id IN (3,4,5)
)
DELETE FROM tournaments_teams_link
WHERE tournaments_teams_link.tournament_id IN (SELECT id FROM tournaments_ids);

-- Удаление игр
WITH stage_ids AS (
    SELECT id FROM tournament_stages
    WHERE tournament_id IN (SELECT id FROM (
        SELECT id FROM tournaments WHERE type_id IN (3,4,5)
    ) AS tournaments_ids)
)
DELETE FROM games
WHERE stage_id IN (SELECT id FROM stage_ids);

-- Удаление стадий турниров
WITH tournaments_ids AS (
    SELECT id FROM tournaments
    WHERE type_id IN (3,4,5)
),
stage_ids AS (
    SELECT id FROM tournament_stages
    WHERE tournament_id IN (SELECT id FROM tournaments_ids)
)
DELETE FROM tournament_stages
WHERE id IN (SELECT id FROM stage_ids);

-- Удаление турниров
DELETE FROM tournaments
WHERE type_id IN (3,4,5);

-- Удаление типов турниров
DELETE FROM tournament_types
WHERE id IN (3,4,5);

-- Удаление специфичных объектов
DROP FUNCTION IF EXISTS get_next_name_suffix;

DROP SEQUENCE IF EXISTS increment_bigint;

COMMIT;
*/