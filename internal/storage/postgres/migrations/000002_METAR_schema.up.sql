BEGIN;

DROP INDEX IF EXISTS metars_issued_at_index;
DROP INDEX IF EXISTS metars_station_id_index;
DROP TABLE IF EXISTS metars;

CREATE TABLE metars (
    metar_id uuid NOT NULL,
    station_id text NOT NULL,
    issued_at bigint NOT NULL,
    raw text NOT NULL,
    PRIMARY KEY (metar_id),
    UNIQUE (station_id, issued_at, raw)
);

CREATE INDEX metars_station_id_index ON metars USING HASH (station_id);
CREATE INDEX metars_issued_at_index ON metars (issued_at);

COMMIT;