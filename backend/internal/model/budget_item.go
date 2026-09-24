package model

import (
	"time"

	"github.com/renovation/renovation-budget-api/internal/constants"
)

// BudgetItem 预算项。
type BudgetItem struct {
	ID             uint                     `gorm:"primaryKey" json:"id"`
	BudgetSheetID  uint                     `gorm:"not null;index" json:"budget_sheet_id"`
	Category       constants.BudgetCategory `gorm:"size:32;not null" json:"category"`
	SubCategory    string                   `gorm:"size:128" json:"sub_category"`
	BudgetAmount   float64                  `gorm:"not null;default:0" json:"budget_amount"`
	SpentAmount    float64                  `gorm:"not null;default:0" json:"spent_amount"`
	VarianceAmount float64                  `gorm:"not null;default:0" json:"variance_amount"`
	SortOrder      int                      `gorm:"not null;default:0" json:"sort_order"`
	Remark         string                   `gorm:"size:512" json:"remark"`
	CreatedAt      time.Time                `json:"created_at"`
	UpdatedAt      time.Time                `json:"updated_at"`
}
