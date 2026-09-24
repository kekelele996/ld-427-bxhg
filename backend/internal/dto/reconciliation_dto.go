package dto

// CreateReconciliationRequest 创建对账单请求。
type CreateReconciliationRequest struct {
	ProjectID     string  `json:"project_id" binding:"required"`
	Period        string  `json:"period" binding:"required,datetime=2006-01"`
	SupplierID    uint    `json:"supplier_id" binding:"required"`
	PayableAmount float64 `json:"payable_amount" binding:"required,gt=0"`
	PaidAmount    float64 `json:"paid_amount" binding:"omitempty,gte=0"`
}

// UpdateReconciliationRequest 更新对账单请求。
type UpdateReconciliationRequest struct {
	PayableAmount *float64 `json:"payable_amount" binding:"omitempty,gt=0"`
	PaidAmount    *float64 `json:"paid_amount" binding:"omitempty,gte=0"`
}

// ReconciliationFilter 对账单列表查询参数。
type ReconciliationFilter struct {
	ProjectID  string `form:"project_id"`
	Period     string `form:"period" binding:"omitempty,datetime=2006-01"`
	SupplierID uint   `form:"supplier_id"`
	Status     string `form:"status" binding:"omitempty,reconciliation_status"`
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// AuditLogFilter 审计日志列表查询参数。
type AuditLogFilter struct {
	Action   string `form:"action"`
	Resource string `form:"resource"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
