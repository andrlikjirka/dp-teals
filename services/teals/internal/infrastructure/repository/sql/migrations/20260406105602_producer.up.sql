CREATE TABLE IF NOT EXISTS teals.producer (
    id          UUID PRIMARY KEY,
    name        VARCHAR NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
