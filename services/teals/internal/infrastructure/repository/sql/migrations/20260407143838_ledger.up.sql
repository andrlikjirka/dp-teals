CREATE TABLE IF NOT EXISTS teals.mmr_node (
    id              BIGSERIAL PRIMARY KEY,
    leaf_index      BIGINT UNIQUE NULL,
    left_child_id   BIGINT NULL REFERENCES teals.mmr_node(id),
    right_child_id  BIGINT NULL REFERENCES teals.mmr_node(id),
    parent_id       BIGINT NULL REFERENCES teals.mmr_node(id),
    hash            BYTEA NOT NULL,
    level           INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_mmr_peaks ON teals.mmr_node(id) WHERE parent_id IS NULL;

ALTER TABLE teals.log_entry ADD COLUMN mmr_node_id BIGINT NULL REFERENCES teals.mmr_node(id);
