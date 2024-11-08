BEGIN;

UPDATE games
    SET date = '2030-01-01' --change this
    WHERE date IS NULL;

ALTER TABLE games
ALTER COLUMN date SET NOT NULL;

COMMIT;