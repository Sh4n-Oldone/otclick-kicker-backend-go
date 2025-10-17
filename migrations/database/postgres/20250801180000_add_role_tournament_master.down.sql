BEGIN;

DROP TABLE IF EXISTS users_cities_links;

DELETE FROM user_roles
WHERE name = 'tournament-master';

COMMIT;