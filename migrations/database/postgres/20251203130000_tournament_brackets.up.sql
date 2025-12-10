BEGIN;

----------------------------------------------------------------
----------------------------------------------------------------

CREATE TABLE IF NOT EXISTS public.tournament_brackets(
    tournament_id INTEGER NOT NULL REFERENCES public.tournaments(id) ON DELETE CASCADE,
    stage_number INTEGER NOT NULL,
    winners_team_id INTEGER REFERENCES public.teams(id) ON DELETE CASCADE,
    losers_team_id INTEGER REFERENCES public.teams(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS tournament_brackets_winners_unique 
    ON public.tournament_brackets(tournament_id, stage_number, winners_team_id)
    WHERE winners_team_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS tournament_brackets_losers_unique 
    ON public.tournament_brackets(tournament_id, stage_number, losers_team_id)
    WHERE losers_team_id IS NOT NULL;
----------------------------------------------------------------
----------------------------------------------------------------

COMMIT;