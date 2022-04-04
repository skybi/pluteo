begin;

drop index if exists metars_issued_at_index;
drop index if exists metars_station_id_index;
drop table if exists metars;

create table metars (
    metar_id uuid not null,
    station_id text not null,
    issued_at bigint not null,
    raw text not null,
    primary key (metar_id),
    unique (station_id, issued_at, raw)
);

create index metars_station_id_index on metars using hash (station_id);
create index metars_issued_at_index on metars (issued_at);

commit;