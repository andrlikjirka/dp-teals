INSERT INTO teals.mmr_node (leaf_index, left_child_id, right_child_id, hash, level)
VALUES ($1, $2, $3, $4, $5)
RETURNING id
