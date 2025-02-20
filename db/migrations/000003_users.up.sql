CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'User ID (Auto-increment primary key)',
    username VARCHAR(50) NOT NULL UNIQUE COMMENT 'Unique username',
    balance DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT 'User balance (default 0)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Account creation timestamp',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Last update timestamp'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User account information';

-- Indexes for efficient queries
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_balance ON users(balance);
