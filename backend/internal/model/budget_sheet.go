package model

import (
	"time"

	"github.com/renovation/renovation-budget-api/internal/constants"
)

// BudgetSheet 预算表。
type BudgetSheet struct {
	ID              uint                   `gorm:"primaryKey" json:"id"`
	ProjectID       string                 `gorm:"size:64;not null;index" json:"project_id"`
	Name            string                 `gorm:"size:128;not null" json:"name"`
	TotalAmount     float64                `gorm:"not null;default:0" json:"total_amount"`
	SpentAmount     float64                `gorm:"not null;default:0" json:"spent_amount"`
	FrozenAmount    float64                `gorm:"not null;default:0" json:"frozen_amount"`
	AvailableAmount float64                `gorm:"not null;default:0" json:"available_amount"`
	Status          constants.BudgetStatus `gorm:"size:32;not null;default:Draft" json:"status"`
	CreatedByID     uint                   `gorm:"not null" json:"created_by_id"`
	ApprovedByID    *uint                  `json:"approved_by_id"`
	Version         int                    `gorm:"not null;default:1" json:"version"`
	Items           []BudgetItem           `gorm:"foreignKey:BudgetSheetID" json:"items,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}
