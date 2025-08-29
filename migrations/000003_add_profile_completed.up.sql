-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN profile_completed BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE users_backup AS SELECT id, email, name, password_hash, created_at FROM users;
DROP TABLE users;
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users (id, email, name, password_hash, created_at) 
SELECT id, email, name, password_hash, created_at FROM users_backup;
DROP TABLE users_backup;
-- +goose StatementEnd
