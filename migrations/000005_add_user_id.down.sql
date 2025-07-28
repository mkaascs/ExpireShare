-- Delete column user_id into files table
-- All data will be deleted nonreturnable. Make back up
ALTER TABLE files DROP CONSTRAINT fk_files_user, DROP COLUMN user_id;