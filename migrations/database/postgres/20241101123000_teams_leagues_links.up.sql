BEGIN;

-- 1. Создать таблицу teams_leagues_links для хранения связей команд с лигами
CREATE TABLE IF NOT EXISTS public.teams_leagues_links (
    team_id INT REFERENCES public.teams(id) ON DELETE CASCADE,
    league_id INT REFERENCES public.leagues(id) ON DELETE CASCADE,
    PRIMARY KEY (team_id, league_id)
);

-- 2. Перенести данные из столбца league_id в teams_leagues_links
INSERT INTO public.teams_leagues_links (team_id, league_id)
SELECT id, league_id FROM public.teams WHERE league_id IS NOT NULL;

-- 3. Удалить столбец league_id из таблицы teams
ALTER TABLE teams DROP COLUMN league_id;

COMMIT;