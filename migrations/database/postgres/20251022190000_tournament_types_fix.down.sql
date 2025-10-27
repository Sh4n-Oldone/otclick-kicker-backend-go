BEGIN;

UPDATE tournament_types
SET name = 'ladder'
WHERE id = 4;

COMMIT;