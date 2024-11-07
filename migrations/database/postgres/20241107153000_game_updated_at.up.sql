BEGIN;

ALTER TABLE games ADD COLUMN updated_at timestamptz default now() not null;

COMMIT;