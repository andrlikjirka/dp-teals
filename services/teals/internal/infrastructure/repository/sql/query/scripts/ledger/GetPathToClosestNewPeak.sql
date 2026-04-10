WITH RECURSIVE path AS (
    SELECT id, hash, parent_id, left_child_id, right_child_id, level
    FROM teals.mmr_node
    WHERE id = $1
    UNION ALL
    SELECT n.id, n.hash, n.parent_id, n.left_child_id, n.right_child_id, n.level
    FROM teals.mmr_node n
             INNER JOIN path p ON n.id = p.parent_id
    WHERE NOT (p.id = ANY($2::bigint[]))
)
SELECT id, hash, parent_id, left_child_id, right_child_id, level
FROM path
ORDER BY level ASC;
