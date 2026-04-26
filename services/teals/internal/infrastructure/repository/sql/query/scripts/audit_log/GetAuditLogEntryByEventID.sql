SELECT
    le.id,
    le.event_id,
    le.mmr_node_id,
    le.producer_key_id,
    le.signature_token,
    le.created_at,
    le.payload,
    mn.leaf_index,
    le.salt
FROM teals.log_entry le
JOIN teals.mmr_node mn ON mn.id = le.mmr_node_id
WHERE le.event_id = $1