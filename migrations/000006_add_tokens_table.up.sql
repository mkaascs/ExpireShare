-- Create tokens table
CREATE TABLE IF NOT EXISTS tokens (
    user_id BIGINT PRIMARY KEY,
    refresh_token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
)