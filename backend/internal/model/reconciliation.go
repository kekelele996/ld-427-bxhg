package model

import (
	"time"

	"github.com/renovation/renovation-budget-api/internal/constants"
)

// Reconciliation 对账单。
type Reconciliation struct {
	ID            uint                           `gorm:"primaryKey" json:"id"`
	ProjectID     string                         `gorm:"size:64;not null;index" json:"project_id"`
	Period        string                         `gorm:"size:32;not null" json:"period"`
	SupplierID    uint                           `gorm:"not null;index" json:"supplier_id"`
	PayableAmount float64                        `gorm:"not null;default:0" json:"payable_amount"`
	PaidAmount    float64                        `gorm:"not null;default:0" json:"paid_amount"`
	UnpaidAmount  float64                        `gorm:"not null;default:0" json:"unpaid_amount"`
	Status        constants.ReconciliationStatus `gorm:"size:32;not null;default:Pending" json:"status"`
	ConfirmedByID *uint                          `json:"confirmed_by_id"`
	CreatedAt     time.Time                      `json:"created_at"`
	UpdatedAt     time.Time                      `json:"updated_at"`
}
