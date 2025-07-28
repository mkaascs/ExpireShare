-- Create tokens table
CREATE TABLE IF NOT EXISTS tokens (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT,
    refresh_token VARCHAR(255) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
)