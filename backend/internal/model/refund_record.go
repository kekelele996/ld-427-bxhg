package model

import "time"

// RefundRecord 退款登记记录。一笔已付款支出可分多次登记退款。
type RefundRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ExpenseID      uint      `gorm:"not null;index" json:"expense_id"`
	Amount         float64   `gorm:"not null" json:"amount"`
	ReceivedDate   time.Time `gorm:"not null" json:"received_date"`
	Reason         string    `gorm:"size:512;not null;default:''" json:"reason"`
	VoucherNo      string    `gorm:"size:128;not null;uniqueIndex" json:"voucher_no"`
	RegisteredByID uint      `gorm:"not null" json:"registered_by_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
