begin;

ALTER TABLE public.matches ADD COLUMN sort SMALLINT;

commit;