-- Add password hast column in database
ALTER TABLE files ADD COLUMN password_hash VARCHAR(255) NULL;