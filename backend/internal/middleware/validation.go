package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/renovation/renovation-budget-api/internal/constants"
)

var validationOnce sync.Once

// ValidationMiddleware 确保自定义枚举校验器已注册，随后交由 DTO tag 完成参数校验。
func ValidationMiddleware() gin.HandlerFunc {
	ConfigureValidation()
	return func(c *gin.Context) {
		c.Next()
	}
}

// ConfigureValidation 注册业务枚举校验器。函数幂等。
//
// 查询过滤 DTO 使用 string 承载枚举值，创建/更新 DTO 使用 constants 中的
// 具名类型承载枚举值。这里同时兼容两种形态，避免合法的列表过滤参数被误判为非法。
func ConfigureValidation() {
	validationOnce.Do(func() {
		engine, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}
		_ = engine.RegisterValidation("budget_status", func(fl validator.FieldLevel) bool {
			switch v := fl.Field().Interface().(type) {
			case constants.BudgetStatus:
				return constants.ValidBudgetStatus(v)
			case string:
				return constants.ValidBudgetStatus(constants.BudgetStatus(v))
			default:
				return false
			}
		})
		_ = engine.RegisterValidation("expense_status", func(fl validator.FieldLevel) bool {
			switch v := fl.Field().Interface().(type) {
			case constants.ExpenseStatus:
				return constants.ValidExpenseStatus(v)
			case string:
				return constants.ValidExpenseStatus(constants.ExpenseStatus(v))
			default:
				return false
			}
		})
		_ = engine.RegisterValidation("payment_method", func(fl validator.FieldLevel) bool {
			switch v := fl.Field().Interface().(type) {
			case constants.PaymentMethod:
				return constants.ValidPaymentMethod(v)
			case string:
				return constants.ValidPaymentMethod(constants.PaymentMethod(v))
			default:
				return false
			}
		})
		_ = engine.RegisterValidation("supplier_status", func(fl validator.FieldLevel) bool {
			switch v := fl.Field().Interface().(type) {
			case constants.SupplierStatus:
				return constants.ValidSupplierStatus(v)
			case string:
				return constants.ValidSupplierStatus(constants.SupplierStatus(v))
			default:
				return false
			}
		})
		_ = engine.RegisterValidation("supplier_category", func(fl validator.FieldLevel) bool {
			switch v := fl.Field().Interface().(type) {
			case constants.SupplierCategory:
				return constants.ValidSupplierCategory(v)
			case string:
				return constants.ValidSupplierCategory(constants.SupplierCategory(v))
			default:
				return false
			}
		})
		_ = engine.RegisterValidation("budget_category", func(fl validator.FieldLevel) bool {
			switch v := fl.Field().Interface().(type) {
			case constants.BudgetCategory:
				return constants.ValidBudgetCategory(v)
			case string:
				return constants.ValidBudgetCategory(constants.BudgetCategory(v))
			default:
				return false
			}
		})
		_ = engine.RegisterValidation("reconciliation_status", func(fl validator.FieldLevel) bool {
			switch v := fl.Field().Interface().(type) {
			case constants.ReconciliationStatus:
				return constants.ValidReconciliationStatus(v)
			case string:
				return constants.ValidReconciliationStatus(constants.ReconciliationStatus(v))
			default:
				return false
			}
		})
	})
}
