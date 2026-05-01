DROP TABLE IF EXISTS teals.subject_secret;

ALTER TABLE teals.log_entry DROP COLUMN IF EXISTS salt;