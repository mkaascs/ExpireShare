-- Add column user_id into files table
ALTER TABLE files
    ADD COLUMN user_id BIGINT,
    ADD CONSTRAINT fk_files_user
    FOREIGN KEY (user_id) REFERENCES users(id);