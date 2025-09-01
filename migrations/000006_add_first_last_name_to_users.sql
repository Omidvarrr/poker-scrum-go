-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN first_name TEXT;
ALTER TABLE users ADD COLUMN last_name TEXT;

UPDATE users SET
    first_name = CASE
        WHEN INSTR(name, ' ') > 0 THEN SUBSTR(name, 1, INSTR(name, ' ') - 1)
        ELSE name
    END,
    last_name = CASE
        WHEN INSTR(name, ' ') > 0 THEN SUBSTR(name, INSTR(name, ' ') + 1)
        ELSE ''
    END;

ALTER TABLE users DROP COLUMN name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN name TEXT;

UPDATE users SET
    name = first_name || ' ' || last_name;

ALTER TABLE users DROP COLUMN first_name;
ALTER TABLE users DROP COLUMN last_name;
-- +goose StatementEnd
