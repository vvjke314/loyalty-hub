-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS withdrawals(
    id uuid PRIMARY KEY,
    order_id TEXT NOT NULL,
    user_id uuid NOT NULL,
    amount NUMERIC(12,2) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_order_id 
        FOREIGN KEY (order_id)
        REFERENCES orders(number)
        ON DELETE CASCADE,
    CONSTRAINT fk_user_id
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd
