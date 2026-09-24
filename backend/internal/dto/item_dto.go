package dto

import "github.com/renovation/renovation-budget-api/internal/constants"

// CreateItemRequest 创建预算项请求。
type CreateItemRequest struct {
	Category     constants.BudgetCategory `json:"category" binding:"required,budget_category"`
	SubCategory  string                   `json:"sub_category" binding:"omitempty,max=128"`
	BudgetAmount float64                  `json:"budget_amount" binding:"required,gt=0"`
	SortOrder    int                      `json:"sort_order" binding:"omitempty,gte=0"`
	Remark       string                   `json:"remark" binding:"omitempty,max=512"`
}

// UpdateItemRequest 更新预算项请求。
type UpdateItemRequest struct {
	Category     constants.BudgetCategory `json:"category" binding:"omitempty,budget_category"`
	SubCategory  string                   `json:"sub_category" binding:"omitempty,max=128"`
	BudgetAmount float64                  `json:"budget_amount" binding:"omitempty,gt=0"`
	SortOrder    *int                     `json:"sort_order" binding:"omitempty,gte=0"`
	Remark       string                   `json:"remark" binding:"omitempty,max=512"`
}
