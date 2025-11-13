BEGIN;

CREATE TABLE tournament_team_extra_points (
   id SERIAL PRIMARY KEY,
   team_id INT NOT NULL REFERENCES teams(id),
   tournament_id INT NOT NULL REFERENCES tournaments(id),
   reason VARCHAR NOT NULL,
   points INT NOT NULL,
   created_at TIMESTAMP DEFAULT NOW(),
   updated_at TIMESTAMP DEFAULT NULL
);

COMMIT;