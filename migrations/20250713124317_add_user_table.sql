-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    login TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    balance NUMERIC(12,2) DEFAULT 0 NOT NULL,
    withdrawn NUMERIC(12,2) DEFAULT 0 NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
