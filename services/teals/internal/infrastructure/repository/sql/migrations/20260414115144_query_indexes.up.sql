CREATE INDEX IF NOT EXISTS idx_log_entry_payload_gin
    ON teals.log_entry USING GIN (payload jsonb_path_ops);

CREATE INDEX IF NOT EXISTS idx_log_entry_payload_timestamp_btree
    ON teals.log_entry ((payload->>'timestamp'));

CREATE INDEX IF NOT EXISTS idx_log_entry_payload_status
    ON teals.log_entry ((payload->'result'->>'status'));

CREATE INDEX IF NOT EXISTS idx_log_entry_payload_action_type
    ON teals.log_entry ((payload->'action'->>'type'));

CREATE INDEX IF NOT EXISTS idx_log_entry_payload_actor_type
    ON teals.log_entry ((payload->'actor'->>'type'));