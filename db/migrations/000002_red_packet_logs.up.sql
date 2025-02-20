CREATE TABLE red_packet_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT 'User ID who grabbed the red packet',
    red_packet_id BIGINT NOT NULL COMMENT 'Red packet ID',
    amount DECIMAL(10,2) NOT NULL COMMENT 'Amount received from the red packet',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when the red packet was grabbed',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Last update timestamp'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Transaction log of grabbed red packets';

-- Indexes for performance optimization
CREATE INDEX idx_red_packet_logs_red_packet_id ON red_packet_logs (red_packet_id);
CREATE INDEX idx_red_packet_logs_user_id ON red_packet_logs (user_id);
