package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/renovation/renovation-budget-api/internal/config"
	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/handler"
	"github.com/renovation/renovation-budget-api/internal/middleware"
	"github.com/renovation/renovation-budget-api/internal/repository"
	"github.com/renovation/renovation-budget-api/internal/response"
)

// Dependencies 路由装配所需依赖。
type Dependencies struct {
	Config                *config.Config
	Redis                 *redis.Client
	Logger                *slog.Logger
	AuditRepo             repository.AuditRepository
	AuthHandler           *handler.AuthHandler
	AuditHandler          *handler.AuditHandler
	BudgetHandler         *handler.BudgetHandler
	ItemHandler           *handler.ItemHandler
	ExpenseHandler        *handler.ExpenseHandler
	SupplierHandler       *handler.SupplierHandler
	ReconciliationHandler *handler.ReconciliationHandler
}

// New 装配中间件与路由。
func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.ErrorHandlerMiddleware())
	engine.Use(middleware.RequestLoggerMiddleware(deps.Logger))
	engine.Use(middleware.ValidationMiddleware())

	engine.GET("/healthz", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "ok"})
	})

	v1 := engine.Group("/api/v1")
	window := time.Duration(deps.Config.RateLimitWindowSeconds) * time.Second
	v1.Use(middleware.RateLimitMiddleware(deps.Redis, deps.Config.RateLimitMax, window))
	v1.POST("/auth/login", deps.AuthHandler.Login)

	authed := v1.Group("")
	authed.Use(middleware.JWTAuthMiddleware(deps.Config.JWTSecret))
	authed.Use(middleware.AuditLogMiddleware(deps.AuditRepo, deps.Logger))

	budgets := authed.Group("/budgets")
	{
		budgets.GET("", middleware.RBACMiddleware(middleware.PermissionView), deps.BudgetHandler.List)
		budgets.POST("", middleware.RBACMiddleware(middleware.PermissionBudgetWrite), deps.BudgetHandler.Create)
		budgets.GET("/:id", middleware.RBACMiddleware(middleware.PermissionView), deps.BudgetHandler.Get)
		budgets.PUT("/:id", middleware.RBACMiddleware(middleware.PermissionBudgetWrite), deps.BudgetHandler.Update)
		budgets.DELETE("/:id", middleware.RBACMiddleware(middleware.PermissionBudgetWrite), deps.BudgetHandler.Delete)
		budgets.POST("/:id/adjust", middleware.RBACMiddleware(middleware.PermissionBudgetWrite), deps.BudgetHandler.Adjust)
		budgets.GET("/:id/items", middleware.RBACMiddleware(middleware.PermissionView), deps.ItemHandler.List)
		budgets.POST("/:id/items", middleware.RBACMiddleware(middleware.PermissionBudgetWrite), deps.ItemHandler.Create)
		budgets.PUT("/:id/items/:item_id", middleware.RBACMiddleware(middleware.PermissionBudgetWrite), deps.ItemHandler.Update)
		budgets.DELETE("/:id/items/:item_id", middleware.RBACMiddleware(middleware.PermissionBudgetWrite), deps.ItemHandler.Delete)
	}

	expenses := authed.Group("/expenses")
	{
		expenses.GET("", middleware.RBACMiddleware(middleware.PermissionView), deps.ExpenseHandler.List)
		expenses.POST("", middleware.RBACMiddleware(middleware.PermissionExpenseCreate), deps.ExpenseHandler.Create)
		expenses.GET("/:id", middleware.RBACMiddleware(middleware.PermissionView), deps.ExpenseHandler.Get)
		expenses.POST("/:id/submit", middleware.RBACMiddleware(middleware.PermissionExpenseCreate), deps.ExpenseHandler.Submit)
		expenses.POST("/:id/approve", middleware.RBACMiddleware(middleware.PermissionExpenseApprove), deps.ExpenseHandler.Approve)
		expenses.POST("/:id/reject", middleware.RBACMiddleware(middleware.PermissionExpenseApprove), deps.ExpenseHandler.Reject)
		expenses.POST("/:id/pay", middleware.RBACMiddleware(middleware.PermissionExpensePay), deps.ExpenseHandler.Pay)
	}

	suppliers := authed.Group("/suppliers")
	{
		suppliers.GET("", middleware.RBACMiddleware(middleware.PermissionView), deps.SupplierHandler.List)
		suppliers.POST("", middleware.RBACMiddleware(middleware.PermissionSupplierWrite), deps.SupplierHandler.Create)
		suppliers.GET("/:id", middleware.RBACMiddleware(middleware.PermissionView), deps.SupplierHandler.Get)
		suppliers.PUT("/:id", middleware.RBACMiddleware(middleware.PermissionSupplierWrite), deps.SupplierHandler.Update)
		suppliers.DELETE("/:id", middleware.RBACMiddleware(middleware.PermissionSupplierWrite), deps.SupplierHandler.Delete)
	}

	reconciliations := authed.Group("/reconciliations")
	{
		reconciliations.GET("", middleware.RBACMiddleware(middleware.PermissionView), deps.ReconciliationHandler.List)
		reconciliations.POST("", middleware.RBACMiddleware(middleware.PermissionReconciliationWrite), deps.ReconciliationHandler.Create)
		reconciliations.GET("/:id", middleware.RBACMiddleware(middleware.PermissionView), deps.ReconciliationHandler.Get)
		reconciliations.PUT("/:id", middleware.RBACMiddleware(middleware.PermissionReconciliationWrite), deps.ReconciliationHandler.Update)
		reconciliations.POST("/:id/confirm", middleware.RBACMiddleware(middleware.PermissionReconciliationConfirm), deps.ReconciliationHandler.Confirm)
		reconciliations.POST("/:id/dispute", middleware.RBACMiddleware(middleware.PermissionReconciliationWrite), deps.ReconciliationHandler.Dispute)
		reconciliations.POST("/:id/resolve", middleware.RBACMiddleware(middleware.PermissionReconciliationConfirm), deps.ReconciliationHandler.Resolve)
		reconciliations.DELETE("/:id", middleware.RBACMiddleware(middleware.PermissionReconciliationWrite), deps.ReconciliationHandler.Delete)
	}

	audits := authed.Group("/audit-logs")
	{
		audits.GET("", middleware.RBACMiddleware(middleware.PermissionAuditView), deps.AuditHandler.List)
	}

	engine.NoRoute(func(c *gin.Context) {
		response.Abort(c, http.StatusNotFound, constants.CodeNotFound, "route not found")
	})
	return engine
}
