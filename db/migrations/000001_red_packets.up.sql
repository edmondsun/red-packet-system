CREATE TABLE red_packets (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    total_amount DECIMAL(10,2) NOT NULL COMMENT 'Total red packet amount',
    remaining_amount DECIMAL(10,2) NOT NULL COMMENT 'Remaining amount in the red packet',
    total_count INT NOT NULL COMMENT 'Total number of red packets',
    remaining_count INT NOT NULL COMMENT 'Remaining red packets count',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1: Available, 0: Expired',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Record creation timestamp',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Last update timestamp'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Red packet table';
