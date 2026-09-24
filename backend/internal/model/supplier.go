package model

import (
	"time"

	"github.com/renovation/renovation-budget-api/internal/constants"
)

// Supplier 供应商。
type Supplier struct {
	ID          uint                       `gorm:"primaryKey" json:"id"`
	Name        string                     `gorm:"size:128;not null;index" json:"name"`
	Category    constants.SupplierCategory `gorm:"size:32;not null" json:"category"`
	Contact     string                     `gorm:"size:64" json:"contact"`
	Phone       string                     `gorm:"size:32" json:"phone"`
	Address     string                     `gorm:"size:255" json:"address"`
	BankName    string                     `gorm:"size:128" json:"bank_name"`
	BankAccount string                     `gorm:"size:128" json:"bank_account"`
	Status      constants.SupplierStatus   `gorm:"size:32;not null;default:Active" json:"status"`
	Rating      int                        `gorm:"not null;default:0" json:"rating"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}
