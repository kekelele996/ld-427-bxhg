DROP TABLE IF EXISTS refund_records;
ALTER TABLE expense_records DROP COLUMN IF EXISTS refunded_amount;
