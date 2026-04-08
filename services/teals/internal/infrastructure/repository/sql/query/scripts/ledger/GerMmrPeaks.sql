SELECT id, hash, level, leaf_index, left_child_id, right_child_id, parent_id
FROM teals.mmr_node
WHERE parent_id IS NULL
ORDER BY id ASC
