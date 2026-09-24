package dto

import "github.com/renovation/renovation-budget-api/internal/constants"

// CreateSupplierRequest 创建供应商请求。
type CreateSupplierRequest struct {
	Name        string                     `json:"name" binding:"required,max=128"`
	Category    constants.SupplierCategory `json:"category" binding:"required,supplier_category"`
	Contact     string                     `json:"contact" binding:"omitempty,max=64"`
	Phone       string                     `json:"phone" binding:"omitempty,max=32"`
	Address     string                     `json:"address" binding:"omitempty,max=255"`
	BankName    string                     `json:"bank_name" binding:"omitempty,max=128"`
	BankAccount string                     `json:"bank_account" binding:"omitempty,max=128"`
	Status      constants.SupplierStatus   `json:"status" binding:"omitempty,supplier_status"`
	Rating      int                        `json:"rating" binding:"omitempty,gte=1,lte=5"`
}

// UpdateSupplierRequest 更新供应商请求。
type UpdateSupplierRequest struct {
	Name        string                     `json:"name" binding:"omitempty,max=128"`
	Category    constants.SupplierCategory `json:"category" binding:"omitempty,supplier_category"`
	Contact     string                     `json:"contact" binding:"omitempty,max=64"`
	Phone       string                     `json:"phone" binding:"omitempty,max=32"`
	Address     string                     `json:"address" binding:"omitempty,max=255"`
	BankName    string                     `json:"bank_name" binding:"omitempty,max=128"`
	BankAccount string                     `json:"bank_account" binding:"omitempty,max=128"`
	Status      constants.SupplierStatus   `json:"status" binding:"omitempty,supplier_status"`
	Rating      int                        `json:"rating" binding:"omitempty,gte=1,lte=5"`
}

// SupplierFilter 供应商列表查询参数。
type SupplierFilter struct {
	Status   string `form:"status" binding:"omitempty,supplier_status"`
	Category string `form:"category" binding:"omitempty,supplier_category"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
