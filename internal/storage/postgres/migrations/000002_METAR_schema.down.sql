BEGIN;

DROP INDEX IF EXISTS metars_issued_at_index;
DROP INDEX IF EXISTS metars_station_id_index;
DROP TABLE IF EXISTS metars;

COMMIT;