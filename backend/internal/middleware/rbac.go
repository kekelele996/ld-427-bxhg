package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/response"
)

// Permission RBAC 权限点。
type Permission string

const (
	PermissionView                  Permission = "view"
	PermissionBudgetWrite           Permission = "budget:write"
	PermissionExpenseCreate         Permission = "expense:create"
	PermissionExpenseApprove        Permission = "expense:approve"
	PermissionExpensePay            Permission = "expense:pay"
	PermissionSupplierWrite         Permission = "supplier:write"
	PermissionReconciliationWrite   Permission = "reconciliation:write"
	PermissionReconciliationConfirm Permission = "reconciliation:confirm"
	PermissionAuditView             Permission = "audit:view"
)

// RolePermissions 角色权限映射。
var RolePermissions = map[constants.RoleName]map[Permission]struct{}{
	constants.RoleAdmin: {
		PermissionView: {}, PermissionBudgetWrite: {}, PermissionExpenseCreate: {}, PermissionExpenseApprove: {}, PermissionExpensePay: {}, PermissionSupplierWrite: {}, PermissionReconciliationWrite: {}, PermissionReconciliationConfirm: {}, PermissionAuditView: {},
	},
	constants.RoleFinanceManager: {
		PermissionView: {}, PermissionBudgetWrite: {}, PermissionExpenseCreate: {}, PermissionExpenseApprove: {}, PermissionExpensePay: {}, PermissionSupplierWrite: {}, PermissionReconciliationWrite: {}, PermissionReconciliationConfirm: {},
	},
	constants.RoleProjectManager: {
		PermissionView: {}, PermissionBudgetWrite: {}, PermissionExpenseCreate: {}, PermissionSupplierWrite: {}, PermissionReconciliationWrite: {},
	},
	constants.RoleAccountant: {
		PermissionView: {}, PermissionExpenseCreate: {},
	},
	constants.RoleOwner: {
		PermissionView: {},
	},
}

// RBACMiddleware 校验当前角色是否拥有指定权限。
func RBACMiddleware(required Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, ok := CurrentActor(c)
		if !ok {
			response.Abort(c, http.StatusUnauthorized, constants.CodeUnauthorized, "unauthenticated")
			return
		}
		permissions, ok := RolePermissions[actor.Role]
		if !ok {
			response.Abort(c, http.StatusForbidden, constants.CodeForbidden, "unknown role")
			return
		}
		if _, ok := permissions[required]; !ok {
			response.Abort(c, http.StatusForbidden, constants.CodeForbidden, "forbidden: insufficient role permission")
			return
		}
		c.Next()
	}
}
