CREATE TABLE IF NOT EXISTS refund_records (
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
CREATE INDEX IF NOT EXISTS idx_refund_records_expense_id ON refund_records(expense_id);
-- 退款凭据号全局唯一，重复凭据只有第一次登记入账。
CREATE UNIQUE INDEX IF NOT EXISTS idx_refund_records_voucher_no ON refund_records(voucher_no);

ALTER TABLE expense_records ADD COLUMN IF NOT EXISTS refunded_amount DOUBLE PRECISION NOT NULL DEFAULT 0;
