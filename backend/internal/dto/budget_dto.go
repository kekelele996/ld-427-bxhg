package dto

import "github.com/renovation/renovation-budget-api/internal/constants"

// CreateBudgetRequest 创建预算表请求。
type CreateBudgetRequest struct {
	ProjectID   string                 `json:"project_id" binding:"required"`
	Name        string                 `json:"name" binding:"required,max=128"`
	TotalAmount float64                `json:"total_amount" binding:"required,gt=0"`
	Status      constants.BudgetStatus `json:"status" binding:"omitempty,budget_status"`
}

// UpdateBudgetRequest 更新预算表请求。
type UpdateBudgetRequest struct {
	Name        string                 `json:"name" binding:"omitempty,max=128"`
	TotalAmount float64                `json:"total_amount" binding:"omitempty,gte=0"`
	Status      constants.BudgetStatus `json:"status" binding:"omitempty,budget_status"`
}

// AdjustBudgetRequest 预算调整请求。
type AdjustBudgetRequest struct {
	TotalAmount float64 `json:"total_amount" binding:"required,gt=0"`
	Reason      string  `json:"reason" binding:"required"`
}
