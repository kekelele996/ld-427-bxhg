package dto

import "github.com/renovation/renovation-budget-api/internal/constants"

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserInfo 登录响应中的用户信息。
type UserInfo struct {
	ID          uint               `json:"id"`
	Username    string             `json:"username"`
	DisplayName string             `json:"display_name"`
	Role        constants.RoleName `json:"role"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	Token     string   `json:"token"`
	ExpiresIn int      `json:"expires_in"`
	User      UserInfo `json:"user"`
}
