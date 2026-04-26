INSERT INTO teals.subject_secret (subject_id, secret)
VALUES ($1, $2)
ON CONFLICT (subject_id) DO NOTHING
RETURNING secret;
