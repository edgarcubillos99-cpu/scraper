-- +goose Up
ALTER TABLE records
    MODIFY client_id VARCHAR(100),
    MODIFY client VARCHAR(100),
    MODIFY date VARCHAR(100),
    MODIFY type VARCHAR(100),
    MODIFY amount VARCHAR(100),
    MODIFY agent VARCHAR(100);

-- +goose Down
ALTER TABLE records
    MODIFY client_id TEXT,
    MODIFY client TEXT,
    MODIFY date TEXT,
    MODIFY type TEXT,
    MODIFY amount TEXT,
    MODIFY agent TEXT;
