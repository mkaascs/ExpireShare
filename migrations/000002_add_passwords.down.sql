-- Remove password hash column from database
-- All data will be deleted nonreturnable. Make back up
ALTER TABLE files DROP COLUMN password_hash;