SELECT id, size, root_hash, anchored_at, kid, signature_token
FROM teals.checkpoint
ORDER BY anchored_at DESC
LIMIT 1;