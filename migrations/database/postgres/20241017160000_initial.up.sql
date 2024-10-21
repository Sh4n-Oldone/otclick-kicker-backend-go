begin;

CREATE TABLE IF NOT EXISTS cities (
    id SERIAL PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL,
    ru VARCHAR UNIQUE NOT NULL,
    deleted_at timestamptz DEFAULT NULL
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
    password VARCHAR NOT NULL,
    role_id smallint references user_roles(id) not null,
    team_id int references teams(id) default null
);

commit;