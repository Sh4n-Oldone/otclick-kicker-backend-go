BEGIN;

CREATE TABLE IF NOT EXISTS tournament_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR,
    description VARCHAR,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

INSERT INTO tournament_types (id, name, description)
VALUES
(1, 'regular', 'все играют со всеми'),
(2, 'playoff', 'игры на выбывание');

CREATE TABLE IF NOT EXISTS tournaments (
    id SERIAL PRIMARY KEY,
    type_id INTEGER REFERENCES tournament_types(id) ON DELETE CASCADE,
    name VARCHAR,
    rules JSONB NOT NULL,
    city_id INTEGER REFERENCES cities(id) ON DELETE CASCADE,
    season_id INTEGER REFERENCES seasons(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS tournament_stages (
    id SERIAL PRIMARY KEY,
    tournament_id INTEGER REFERENCES tournaments(id) ON DELETE CASCADE,
    is_finished BOOLEAN
);

CREATE TABLE IF NOT EXISTS tournaments_teams_link (
    team_id INTEGER REFERENCES teams(id),
    tournament_id INTEGER REFERENCES tournaments(id)
);

ALTER TABLE games
ADD COLUMN stage_id INTEGER REFERENCES tournament_stages(id);

CREATE TABLE IF NOT EXISTS tournament_rating(
    player_id INTEGER REFERENCES players(id),
    tournament_id INTEGER REFERENCES tournaments(id),
    value INTEGER DEFAULT 1000 CHECK (value BETWEEN 0 AND 10000),
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY(player_id, tournament_id)
);

COMMIT;