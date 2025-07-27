-- Add user id column to files table
ALTER TABLE files ADD COLUMN user_id BIGINT NOT NULL;