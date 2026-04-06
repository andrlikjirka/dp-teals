CREATE TYPE teals.producer_key_status AS ENUM ('active', 'revoked');

CREATE TABLE IF NOT EXISTS teals.producer_key (
    id          UUID PRIMARY KEY,
    producer_id UUID NOT NULL REFERENCES teals.producer (id) ON DELETE RESTRICT,
    kid         TEXT NOT NULL UNIQUE ,
    public_key  bytea NOT NULL,
    status      teals.producer_key_status NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);