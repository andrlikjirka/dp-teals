WITH RECURSIVE path AS (
    SELECT id, hash, parent_id, left_child_id, right_child_id, level, leaf_index
    FROM teals.mmr_node
    WHERE leaf_index = $1
    UNION ALL
    SELECT n.id, n.hash, n.parent_id, n.left_child_id, n.right_child_id, n.level, n.leaf_index
    FROM teals.mmr_node n
             INNER JOIN path ON n.id = path.parent_id
)
SELECT id, hash, parent_id, left_child_id, right_child_id, level, leaf_index
FROM path
ORDER BY level ASC;