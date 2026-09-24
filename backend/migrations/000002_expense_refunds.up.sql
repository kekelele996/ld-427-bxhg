CREATE TABLE IF NOT EXISTS expense_refunds (
    id BIGSERIAL PRIMARY KEY,
    expense_id BIGINT NOT NULL REFERENCES expense_records(id),
    amount DOUBLE PRECISION NOT NULL,
    received_date DATE NOT NULL,
    reason VARCHAR(512) NOT NULL DEFAULT '',
    voucher_no VARCHAR(128) NOT NULL,
    registered_by_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_expense_refunds_expense_id ON expense_refunds(expense_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_expense_refunds_voucher_no ON expense_refunds(voucher_no);
