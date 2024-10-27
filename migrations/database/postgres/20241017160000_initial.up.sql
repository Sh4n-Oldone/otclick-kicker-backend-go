begin;

CREATE TABLE IF NOT EXISTS cities (
    id SERIAL PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL,
    ru VARCHAR UNIQUE NOT NULL,
    deleted_at timestamptz DEFAULT NULL
);

INSERT INTO cities (name, ru)
values
    ('St.Peterburg', 'Санкт-Петербург'),
    ('Moscow', 'Москва'),
    ('Voronezh', 'Воронеж');

CREATE TABLE IF NOT EXISTS tables (
    id SMALLSERIAL PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE,
    updated_at timestamptz default now() not null,
    deleted_at timestamptz default null
);

INSERT INTO tables (name)
VALUES
    ('Garlando World Champion'),
    ('Leo Tournament'),
    ('Tornado T3000'),
    ('Leonhart');

CREATE TABLE IF NOT EXISTS bars (
    id SMALLSERIAL PRIMARY KEY,
    city_id int references cities(id) not null,
    name VARCHAR NOT NULL UNIQUE,
    description VARCHAR,
    updated_at timestamptz default now() not null,
    deleted_at timestamptz default null
);

CREATE TABLE IF NOT EXISTS places (
    id SMALLSERIAL PRIMARY KEY,
    bar_id int references bars(id) not null,
    table_id int references tables(id) not null,
    updated_at timestamptz default now() not null,
    deleted_at timestamptz default null
);

CREATE TABLE IF NOT EXISTS leagues (
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL,
    city_id INT REFERENCES cities(id)
);

CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL,
    short_name VARCHAR UNIQUE NOT NULL,
    league_id INT REFERENCES leagues(id),
    city_id INT REFERENCES cities(id),
    avatar BYTEA
);

CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,
    name VARCHAR,
    second_name VARCHAR,
    last_name VARCHAR NOT NULL,
    active_player BOOLEAN DEFAULT true,
    deleted_at timestamptz DEFAULT null,
    avatar BYTEA,
    city_id INT REFERENCES cities(id)
);

CREATE TABLE IF NOT EXISTS rating (
    player_id int references players(id),
    league_id int references leagues(id),
    value INT DEFAULT 1000 CHECK (value BETWEEN 0 AND 10000)
);

CREATE TABLE IF NOT EXISTS players_teams_links (
    player_id int references players(id),
    team_id int references teams(id),
    primary key (player_id, team_id)
);

CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    city_id INT REFERENCES cities(id),
    place_id INT REFERENCES places(id),
    date date NOT NULL,
    team1_id INT REFERENCES teams(id),
    team2_id INT REFERENCES teams(id)
);

CREATE TABLE IF NOT EXISTS matches (
    id SERIAL PRIMARY KEY,
    date date NOT NULL,
    game_id INT REFERENCES games(id),
    team1_id INT REFERENCES teams(id),
    team2_id INT REFERENCES teams(id),
    player1_team1_id INT REFERENCES players(id),
    player2_team1_id INT REFERENCES players(id),
    player1_team2_id INT REFERENCES players(id),
    player2_team2_id INT REFERENCES players(id),
    score_team1 INT DEFAULT 0,
    score_team2 INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_roles (
    id SMALLSERIAL PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE,
    description VARCHAR,
    updated_at timestamptz default now() not null
);

insert into user_roles (name, description)
values
    ('superuser', 'superuser'),
    ('admin', 'admin'),
    ('captain', 'captain');

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR NOT NULL UNIQUE,
    password BYTEA NOT NULL,
    role_id smallint references user_roles(id) not null,
    team_id int references teams(id) default null
);

commit;