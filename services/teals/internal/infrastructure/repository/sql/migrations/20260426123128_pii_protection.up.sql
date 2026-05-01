CREATE TABLE IF NOT EXISTS teals.subject_secret (
    id          BIGSERIAL PRIMARY KEY,
    subject_id  TEXT NOT NULL UNIQUE,
    secret      BYTEA NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE teals.log_entry ADD COLUMN salt BYTEA NULL;
