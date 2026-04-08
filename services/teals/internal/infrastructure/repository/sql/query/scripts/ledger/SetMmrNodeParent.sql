UPDATE teals.mmr_node
SET parent_id = $1
WHERE id IN ($2, $3)
