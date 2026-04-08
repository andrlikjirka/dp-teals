SELECT id, leaf_index, left_child_id, right_child_id, parent_id, hash, level
FROM teals.mmr_node
WHERE parent_id IS NULL
  AND level = $1
  AND id != $2
LIMIT 1
