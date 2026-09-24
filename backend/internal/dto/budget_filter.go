package dto

// BudgetFilter 预算表列表查询参数。
type BudgetFilter struct {
	ProjectID string `form:"project_id"`
	Status    string `form:"status" binding:"omitempty,budget_status"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
