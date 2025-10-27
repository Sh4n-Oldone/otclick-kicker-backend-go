BEGIN;

UPDATE tournament_types
SET name = 'regular/playoff with loser bracket'
WHERE id = 4;

COMMIT;