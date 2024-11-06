BEGIN;

-- 1. Добавить столбец league_id обратно в таблицу teams
ALTER TABLE public.teams ADD COLUMN league_id INT REFERENCES public.leagues(id);

-- 2. Перенести данные из таблицы teams_leagues_links в teams.league_id
-- Здесь мы предполагаем, что у каждой команды может быть только одна лига
UPDATE public.teams t
SET league_id = (
    SELECT league_id 
    FROM public.teams_leagues_links tll 
    WHERE tll.team_id = t.id
)
WHERE EXISTS (
    SELECT 1 
    FROM public.teams_leagues_links tll 
    WHERE tll.team_id = t.id
);

-- 3. Удалить таблицу teams_leagues_links
DROP TABLE IF EXISTS public.teams_leagues_links;

COMMIT;
