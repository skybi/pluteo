BEGIN;

DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS user_api_key_policies;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    user_id text NOT NULL,
    display_name text NOT NULL DEFAULT '',
    restricted boolean NOT NULL DEFAULT false,
    admin boolean NOT NULL DEFAULT false,
    PRIMARY KEY (user_id)
);

CREATE TABLE user_api_key_policies (
    user_id text NOT NULL,
    max_quota bigint NOT NULL DEFAULT -1,
    max_rate_limit int NOT NULL DEFAULT -1,
    allowed_capabilities int NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE api_keys (
    key_id uuid NOT NULL,
    api_key bytea NOT NULL,
    user_id text NOT NULL,
    description text NOT NULL,
    quota bigint NOT NULL DEFAULT -1,
    used_quota bigint NOT NULL DEFAULT 0,
    rate_limit int NOT NULL DEFAULT -1,
    capabilities int NOT NULL DEFAULT 0,
    PRIMARY KEY (key_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

COMMIT;