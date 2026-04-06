-- Migration script for creating audit_events table

CREATE TABLE IF NOT EXISTS teals.log_entry (
    id          BIGSERIAL PRIMARY KEY,
    event_id    UUID        NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload     JSONB       NOT NULL
);
