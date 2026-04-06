ALTER TABLE teals.log_entry
    DROP COLUMN IF EXISTS signature_token,
    DROP COLUMN IF EXISTS producer_key_id;