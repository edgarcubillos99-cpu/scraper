-- +goose Up
ALTER TABLE records
    MODIFY `amount` DECIMAL(10,2) NULL;

-- +goose Down
ALTER TABLE records
    MODIFY `amount` VARCHAR(100) NULL;
