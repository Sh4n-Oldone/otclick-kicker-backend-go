-- Для единого сезона для всех городов:

BEGIN;

INSERT INTO seasons(name, description)
VALUES ('Season 2024-2025', 'migrate from cities/years')
ON CONFLICT (name) DO NOTHING;

UPDATE leagues
SET season_id = (SELECT id FROM seasons WHERE name = 'Season 2024-2025')
WHERE season_id IS NULL;

COMMIT;