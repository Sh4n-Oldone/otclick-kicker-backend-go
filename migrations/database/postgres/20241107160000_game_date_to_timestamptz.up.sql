BEGIN;

ALTER TABLE games
ALTER COLUMN date TYPE timestamptz USING date::timestamptz;

ALTER TABLE matches
ALTER COLUMN date TYPE timestamptz USING date::timestamptz;

COMMIT;