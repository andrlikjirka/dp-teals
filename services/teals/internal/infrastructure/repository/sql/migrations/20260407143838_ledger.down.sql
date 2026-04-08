ALTER TABLE teals.log_entry DROP COLUMN IF EXISTS mmr_node_id;

DROP TABLE IF EXISTS teals.mmr_node CASCADE;