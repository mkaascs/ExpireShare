-- Create table for file
CREATE TABLE IF NOT EXISTS files (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    file_path TEXT NOT NULL,
    alias VARCHAR(50) NOT NULL,
    downloads_left SMALLINT,
    loaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    UNIQUE KEY (alias),
    INDEX (alias)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
