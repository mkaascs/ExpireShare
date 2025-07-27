-- All data will be deleted nonreturnable. Make back up
-- Remove user id column from files table
ALTER TABLE files DROP COLUMN user_id;

-- Drop user database
DROP TABLE IF EXISTS users;