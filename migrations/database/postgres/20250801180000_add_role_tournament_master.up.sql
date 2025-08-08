BEGIN;

INSERT INTO user_roles (name, description, updated_at)
VALUES ('tournament-master','tournament master', NOW());

CREATE TABLE IF NOT EXISTS users_cities_links (
    user_id INTEGER REFERENCES users(id),
    city_id INTEGER REFERENCES cities(id),
    PRIMARY KEY (user_id, city_id)
);

COMMIT;