ALTER TABLE teals.log_entry
    ADD COLUMN signature_token TEXT NOT NULL DEFAULT '',
    ADD COLUMN producer_key_id UUID NOT NULL REFERENCES teals.producer_key(id) ON DELETE RESTRICT DEFAULT '00000000-0000-0000-0000-000000000000';

ALTER TABLE teals.log_entry
    ALTER COLUMN signature_token DROP DEFAULT,
    ALTER COLUMN producer_key_id DROP DEFAULT;