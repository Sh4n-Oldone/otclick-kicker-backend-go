BEGIN;

ALTER TABLE games
ALTER COLUMN date TYPE date USING date::date;

ALTER TABLE matches
ALTER COLUMN date TYPE date USING date::date;

COMMIT;