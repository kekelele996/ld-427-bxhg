package repository

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/model"
)

// SeedInitialData 初始化 RBAC 角色与演示用户。该函数由 main 显式调用。
func SeedInitialData(db *gorm.DB) error {
	ctx := context.Background()
	roles := []model.Role{
		{Name: string(constants.RoleAdmin), Description: "系统管理员，拥有全部权限"},
		{Name: string(constants.RoleFinanceManager), Description: "财务经理，可审批支出、付款与对账"},
		{Name: string(constants.RoleProjectManager), Description: "项目经理，可编制预算和提交支出"},
		{Name: string(constants.RoleAccountant), Description: "会计，可录入支出但不能审批"},
		{Name: string(constants.RoleOwner), Description: "业主，仅可查看数据"},
	}
	for i := range roles {
		var existing model.Role
		err := db.WithContext(ctx).Where("name = ?", roles[i].Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.WithContext(ctx).Create(&roles[i]).Error; err != nil {
				return fmt.Errorf("seed role %s: %w", roles[i].Name, err)
			}
		} else if err != nil {
			return fmt.Errorf("query role %s: %w", roles[i].Name, err)
		}
	}

	users := []struct {
		username string
		password string
		role     constants.RoleName
	}{
		{"admin", "Admin123!", constants.RoleAdmin},
		{"finance", "Finance123!", constants.RoleFinanceManager},
		{"project", "Project123!", constants.RoleProjectManager},
		{"accountant", "Accountant123!", constants.RoleAccountant},
		{"owner", "Owner123!", constants.RoleOwner},
	}
	for _, u := range users {
		var role model.Role
		if err := db.WithContext(ctx).Where("name = ?", string(u.role)).First(&role).Error; err != nil {
			return fmt.Errorf("query role for user %s: %w", u.username, err)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", u.username, err)
		}
		user := model.User{Username: u.username, PasswordHash: string(hash), DisplayName: u.username, RoleID: role.ID}
		var existing model.User
		err = db.WithContext(ctx).Where("username = ?", u.username).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.WithContext(ctx).Create(&user).Error; err != nil {
				return fmt.Errorf("seed user %s: %w", u.username, err)
			}
		} else if err != nil {
			return fmt.Errorf("query user %s: %w", u.username, err)
		}
	}
	return nil
}
