-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sessions(
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    hashed_refresh_token TEXT NOT NULL,
    auth_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expire_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT fk_user_id
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
