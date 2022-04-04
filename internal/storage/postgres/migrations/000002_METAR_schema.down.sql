begin;

drop index if exists metars_issued_at_index;
drop index if exists metars_station_id_index;
drop table if exists metars;

commit;