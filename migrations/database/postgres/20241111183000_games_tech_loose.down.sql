BEGIN;

ALTER TABLE public.games DROP COLUMN tech_loose_team_id;

COMMIT;