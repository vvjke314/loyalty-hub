-- +goose Up
-- +goose StatementBegin
CREATE TYPE order_status AS ENUM ('NEW', 'REGISTERED', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS orders(
    number TEXT PRIMARY KEY,
    user_id uuid NOT NULL,
    status order_status NOT NULL,
    accrual NUMERIC(12,2),
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_user_id
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders CASCADE;
DROP TYPE IF EXISTS order_status CASCADE;
-- +goose StatementEnd
