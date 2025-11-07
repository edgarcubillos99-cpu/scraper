-- +goose Up
CREATE INDEX idx_client_id ON records(client_id);
CREATE INDEX idx_client ON records(client);
CREATE INDEX idx_date ON records(date);
CREATE INDEX idx_type ON records(type);
CREATE INDEX idx_amount ON records(amount);
CREATE INDEX idx_agent ON records(agent);

-- +goose Down
DROP INDEX idx_client_id ON records;
DROP INDEX idx_client ON records;
DROP INDEX idx_date ON records;
DROP INDEX idx_type ON records;
DROP INDEX idx_amount ON records;
DROP INDEX idx_agent ON records;

