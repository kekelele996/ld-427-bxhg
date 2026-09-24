package dto

import "github.com/renovation/renovation-budget-api/internal/model"

// CreateRefundRequest 退款登记请求。
type CreateRefundRequest struct {
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	ReceivedDate string  `json:"received_date" binding:"required,datetime=2006-01-02"`
	Reason       string  `json:"reason" binding:"omitempty,max=512"`
	VoucherNo    string  `json:"voucher_no" binding:"required,max=128"`
}

// ExpenseRefundDetail 支出详情中的退款汇总与明细。
type ExpenseRefundDetail struct {
	RefundedAmount float64              `json:"refunded_amount"`
	RefundableLeft float64              `json:"refundable_left"`
	Refunds        []model.RefundRecord `json:"refunds"`
}

// ExpenseDetailResponse 支出详情，包含退款记录、已退金额与剩余可退金额。
type ExpenseDetailResponse struct {
	model.ExpenseRecord
	Refund ExpenseRefundDetail `json:"refund"`
}
