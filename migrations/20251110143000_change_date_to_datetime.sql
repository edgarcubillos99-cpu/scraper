-- +goose Up
ALTER TABLE records
    MODIFY `date` DATETIME(3) NULL;

-- +goose Down
ALTER TABLE records
    MODIFY `date` VARCHAR(100) NULL;
