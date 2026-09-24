package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	_ "github.com/renovation/renovation-budget-api/docs"
	"github.com/renovation/renovation-budget-api/internal/config"
	"github.com/renovation/renovation-budget-api/internal/handler"
	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/repository"
	"github.com/renovation/renovation-budget-api/internal/router"
	"github.com/renovation/renovation-budget-api/internal/service"
)

// @title Renovation Budget API
// @version 1.0.0
// @description 装修预算与支出管理 API 服务，提供预算编制、支出审批、供应商管理与对账结算能力。
// @termsOfService http://example.com/terms/
// @contact.name API Support
// @contact.email support@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:19306
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db, err := initDB(cfg, logger)
	if err != nil {
		logger.Error("init database failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := repository.SeedInitialData(db); err != nil {
		logger.Error("seed initial data failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	rdb, err := initRedis(cfg, logger)
	if err != nil {
		logger.Error("init redis failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	engine := buildEngine(cfg, db, rdb, logger)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: engine,
	}

	go func() {
		logger.Info("server started", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server listen failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
	}
	logger.Info("server stopped")
}

func initDB(cfg *config.Config, logger *slog.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.BudgetSheet{},
		&model.BudgetItem{},
		&model.ExpenseRecord{},
		&model.Supplier{},
		&model.Reconciliation{},
		&model.AuditLog{},
	); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	logger.Info("database migrated")
	return db, nil
}

func initRedis(cfg *config.Config, logger *slog.Logger) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	logger.Info("redis connected")
	return rdb, nil
}

func buildEngine(cfg *config.Config, db *gorm.DB, rdb *redis.Client, logger *slog.Logger) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	budgetRepo := repository.NewBudgetRepository(db)
	itemRepo := repository.NewItemRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	reconciliationRepo := repository.NewReconciliationRepository(db)

	auditService := service.NewAuditService(auditRepo, logger)
	authService := service.NewAuthService(userRepo, cfg, logger)
	budgetService := service.NewBudgetService(budgetRepo, itemRepo, auditService, rdb, logger)
	itemService := service.NewItemService(itemRepo, budgetRepo, auditService, rdb, logger)
	expenseService := service.NewExpenseService(expenseRepo, itemRepo, budgetRepo, auditService, rdb, logger)
	supplierService := service.NewSupplierService(supplierRepo, auditService, logger)
	reconciliationService := service.NewReconciliationService(reconciliationRepo, auditService, logger)

	engine := router.New(router.Dependencies{
		Config:                cfg,
		Redis:                 rdb,
		Logger:                logger,
		AuditRepo:             auditRepo,
		AuthHandler:           handler.NewAuthHandler(authService, logger),
		AuditHandler:          handler.NewAuditHandler(auditService, logger),
		BudgetHandler:         handler.NewBudgetHandler(budgetService, logger),
		ItemHandler:           handler.NewItemHandler(itemService, logger),
		ExpenseHandler:        handler.NewExpenseHandler(expenseService, logger),
		SupplierHandler:       handler.NewSupplierHandler(supplierService, logger),
		ReconciliationHandler: handler.NewReconciliationHandler(reconciliationService, logger),
	})

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return engine
}
