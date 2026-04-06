UPDATE teals.producer_key
SET status = 'revoked'
WHERE kid = $1
