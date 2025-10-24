BEGIN;

CREATE SEQUENCE increment_bigint START 1;

CREATE OR REPLACE FUNCTION get_next_name_suffix()
RETURNS VARCHAR SECURITY DEFINER LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN nextval('increment_bigint')::VARCHAR;
END;
$$;

INSERT INTO tournament_types (id, name, description, updated_at)
VALUES
(3, 'regular/playoff', 'на выбывание с отбором',now()),
(4, 'ladder', 'отбор + на выбывание сетки виннеров и лузеров',now()),
(5, 'regular one vs one', 'каждый с каждым, по одному человеку в команде',now());

COMMIT;