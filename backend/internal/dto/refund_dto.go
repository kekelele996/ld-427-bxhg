package dto

import "github.com/renovation/renovation-budget-api/internal/model"

// RegisterRefundRequest 退款登记请求。
type RegisterRefundRequest struct {
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	ReceivedDate string  `json:"received_date" binding:"required,datetime=2006-01-02"`
	Reason       string  `json:"reason" binding:"required,max=512"`
	VoucherNo    string  `json:"voucher_no" binding:"required,max=128"`
}

// ExpenseRefundSummary 支出退款汇总。
type ExpenseRefundSummary struct {
	Records        []model.ExpenseRefund `json:"records"`
	RefundedAmount float64               `json:"refunded_amount"`
	RefundableLeft float64               `json:"refundable_left"`
}

// ExpenseDetailResponse 支出详情，含退款记录与退款金额汇总。
type ExpenseDetailResponse struct {
	model.ExpenseRecord
	RefundedAmount float64               `json:"refunded_amount"`
	RefundableLeft float64               `json:"refundable_left"`
	Refunds        []model.ExpenseRefund `json:"refunds"`
}
