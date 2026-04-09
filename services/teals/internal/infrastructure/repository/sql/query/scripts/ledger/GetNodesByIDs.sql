SELECT id, hash, parent_id, left_child_id, right_child_id, level, leaf_index
FROM teals.mmr_node
WHERE id = ANY($1)
