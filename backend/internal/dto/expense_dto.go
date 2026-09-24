package dto

import "github.com/renovation/renovation-budget-api/internal/constants"

// CreateExpenseRequest 创建支出记录请求。
type CreateExpenseRequest struct {
	BudgetItemID  uint                    `json:"budget_item_id" binding:"required"`
	Amount        float64                 `json:"amount" binding:"required,gt=0"`
	ExpenseDate   string                  `json:"expense_date" binding:"required,datetime=2006-01-02"`
	PaymentMethod constants.PaymentMethod `json:"payment_method" binding:"required,payment_method"`
	SupplierID    *uint                   `json:"supplier_id"`
	InvoiceNo     string                  `json:"invoice_no" binding:"omitempty,max=128"`
	Description   string                  `json:"description" binding:"omitempty,max=512"`
	AttachmentURL string                  `json:"attachment_url" binding:"omitempty,max=512"`
}

// ApproveExpenseRequest 审批通过支出请求。
type ApproveExpenseRequest struct {
	ApprovalComment string `json:"approval_comment" binding:"omitempty,max=512"`
}

// RejectExpenseRequest 驳回支出请求。
type RejectExpenseRequest struct {
	ApprovalComment string `json:"approval_comment" binding:"required,max=512"`
}

// PayExpenseRequest 付款请求。
type PayExpenseRequest struct {
	PaymentDate string `json:"payment_date" binding:"omitempty,datetime=2006-01-02"`
}
