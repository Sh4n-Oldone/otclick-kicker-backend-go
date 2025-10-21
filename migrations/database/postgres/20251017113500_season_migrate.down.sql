BEGIN;

UPDATE leagues SET season_id = NULL
WHERE season_id = (SELECT id FROM seasons WHERE name = 'Season 2024-2025');

DELETE FROM seasons
WHERE name = 'Season 2024-2025';

COMMIT;