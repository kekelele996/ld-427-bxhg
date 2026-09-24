package dto

// ExpenseFilter 支出记录列表查询参数。
type ExpenseFilter struct {
	Status       string `form:"status" binding:"omitempty,expense_status"`
	BudgetItemID uint   `form:"budget_item_id"`
	SupplierID   uint   `form:"supplier_id"`
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
