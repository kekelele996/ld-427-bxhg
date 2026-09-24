package model

import (
	"time"

	"github.com/renovation/renovation-budget-api/internal/constants"
)

// AuditLog 审计日志。
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	Username   string    `gorm:"size:64" json:"username"`
	Action     string    `gorm:"size:64;not null;index" json:"action"`
	Resource   string    `gorm:"size:64" json:"resource"`
	ResourceID string    `gorm:"size:64;index" json:"resource_id"`
	Detail     string    `gorm:"type:text" json:"detail"`
	IP         string    `gorm:"size:64" json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}

// Actor 请求中的操作者，从 JWT 与请求上下文提取。
type Actor struct {
	UserID   uint
	Username string
	Role     constants.RoleName
	IP       string
}
