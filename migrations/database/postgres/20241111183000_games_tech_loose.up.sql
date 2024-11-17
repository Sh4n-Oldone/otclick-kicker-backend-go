BEGIN;

ALTER TABLE public.games ADD COLUMN tech_loose_team_id INT REFERENCES teams(id);

COMMIT;