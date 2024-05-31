CREATE TABLE IF NOT EXISTS sso.tokens
(
    hash    bytea PRIMARY KEY,
    user_id BIGINT                      NOT NULL REFERENCES sso.users ON DELETE CASCADE ,
    expiry  TIMESTAMP(0) WITH TIME ZONE NOT NULL
);