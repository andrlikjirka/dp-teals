INSERT INTO teals.log_entry (event_id, payload, signature_token, producer_key_id, mmr_node_id, salt)
VALUES ($1, $2, $3, $4, $5, $6);