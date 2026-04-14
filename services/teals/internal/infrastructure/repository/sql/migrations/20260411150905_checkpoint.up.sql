CREATE TABLE IF NOT EXISTS teals.checkpoint (
    id              UUID        PRIMARY KEY,
    size            BIGINT      NOT NULL,
    root_hash       BYTEA       NOT NULL,
    anchored_at     TIMESTAMPTZ NOT NULL,
    kid             TEXT        NOT NULL,
    signature_token TEXT        NOT NULL
);

CREATE INDEX idx_checkpoint_anchored_at ON teals.checkpoint (anchored_at DESC);
