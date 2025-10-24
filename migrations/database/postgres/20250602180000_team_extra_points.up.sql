BEGIN;

CREATE TABLE team_extra_points (
    id SERIAL PRIMARY KEY,
    team_id INT NOT NULL REFERENCES teams(id),
    league_id INT NOT NULL REFERENCES leagues(id),
    reason VARCHAR NOT NULL,
    points INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

COMMIT;