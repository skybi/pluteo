begin;

drop table if exists api_keys;
drop table if exists user_api_key_policies;
drop table if exists users;

create table users (
    user_id text not null,
    display_name text not null default '',
    restricted boolean not null default false,
    admin boolean not null default false,
    primary key (user_id)
);

create table user_api_key_policies (
    user_id text not null,
    max_quota bigint not null default -1,
    max_rate_limit int not null default -1,
    allowed_capabilities int not null default 0,
    primary key (user_id),
    foreign key (user_id) references users(user_id) on delete cascade
);

create table api_keys (
    key_id uuid not null,
    api_key bytea not null,
    user_id text not null,
    description text not null,
    quota bigint not null default -1,
    used_quota bigint not null default 0,
    rate_limit int not null default -1,
    capabilities int not null default 0,
    primary key (key_id),
    foreign key (user_id) references users(user_id) on delete cascade
);

commit;