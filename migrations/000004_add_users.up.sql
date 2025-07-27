--- Add user id column to files table
ALTER TABLE files ADD COLUMN user_id BIGINT NOT NULL;

--- Create user table in database
CREATE TABLE IF NOT EXISTS users(
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    ip VARCHAR(50) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;