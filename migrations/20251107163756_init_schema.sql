-- +goose Up
CREATE TABLE IF NOT EXISTS records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    client_id VARCHAR(50),
    client VARCHAR(255),
    date VARCHAR(50),
    type VARCHAR(100),
    amount VARCHAR(20),
    agent VARCHAR(100)
);

-- +goose Down
DROP TABLE IF EXISTS records;
