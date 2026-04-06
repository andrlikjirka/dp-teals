SELECT id, kid, producer_id, public_key, status, created_at
FROM teals.producer_key
WHERE kid = $1 AND status = $2
