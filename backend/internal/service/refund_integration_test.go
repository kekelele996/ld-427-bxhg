package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/repository"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date %s: %v", s, err)
	}
	return d
}

func newIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.BudgetSheet{},
		&model.BudgetItem{},
		&model.ExpenseRecord{},
		&model.RefundRecord{},
		&model.AuditLog{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

// TestRefundServiceIntegration 基于真实数据库事务验证：分次退款回冲、
// 超额整体回滚余额不变、凭据号重复仅第一次入账。
func TestRefundServiceIntegration(t *testing.T) {
	ctx := context.Background()
	db := newIntegrationDB(t)

	budgetRepo := repository.NewBudgetRepository(db)
	itemRepo := repository.NewItemRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	refundRepo := repository.NewRefundRepository(db)
	audit := NewAuditService(repository.NewAuditRepository(db), testLogger())
	refundSvc := NewRefundService(repository.NewTxManager(db), refundRepo, expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger())

	sheet := &model.BudgetSheet{ProjectID: "p-1", Name: "整屋装修", TotalAmount: 10000, SpentAmount: 1000, AvailableAmount: 9000, Status: constants.BudgetStatusActive, CreatedByID: 1, Version: 1}
	if err := budgetRepo.Create(ctx, sheet); err != nil {
		t.Fatalf("create sheet: %v", err)
	}
	item := &model.BudgetItem{BudgetSheetID: sheet.ID, Category: constants.BudgetCategoryMaterial, BudgetAmount: 5000, SpentAmount: 1000, VarianceAmount: -4000}
	if err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	expense := &model.ExpenseRecord{
		BudgetItemID: item.ID, Amount: 1000, ExpenseDate: mustDate(t, "2026-08-01"),
		PaymentMethod: constants.PaymentMethodBankTransfer, Status: constants.ExpenseStatusPaid,
		ApplicantID: 2,
	}
	if err := expenseRepo.Create(ctx, expense); err != nil {
		t.Fatalf("create expense: %v", err)
	}
	actor := model.Actor{UserID: 3, Username: "finance"}

	if _, err := refundSvc.Register(ctx, actor, expense.ID, dto.CreateRefundRequest{
		Amount: 400, ReceivedDate: "2026-08-15", Reason: "退料", VoucherNo: "INT-0001",
	}); err != nil {
		t.Fatalf("first refund: %v", err)
	}

	// 超额退款：400 + 700 > 1000，事务回滚，三处余额均保持第一次退款后的状态。
	if _, err := refundSvc.Register(ctx, actor, expense.ID, dto.CreateRefundRequest{
		Amount: 700, ReceivedDate: "2026-08-16", Reason: "减项", VoucherNo: "INT-0002",
	}); !errors.Is(err, ErrRefundExceeded) {
		t.Fatalf("error = %v, want ErrRefundExceeded", err)
	}
	gotExpense, _ := expenseRepo.FindByID(ctx, expense.ID)
	if gotExpense.RefundedAmount != 400 {
		t.Fatalf("expense refunded = %v, want 400", gotExpense.RefundedAmount)
	}
	gotItem, _ := itemRepo.FindByID(ctx, item.ID)
	if gotItem.SpentAmount != 600 {
		t.Fatalf("item spent = %v, want 600", gotItem.SpentAmount)
	}
	gotSheet, _ := budgetRepo.FindByID(ctx, sheet.ID)
	if gotSheet.SpentAmount != 600 || gotSheet.AvailableAmount != 9400 {
		t.Fatalf("budget spent=%v available=%v, want 600/9400", gotSheet.SpentAmount, gotSheet.AvailableAmount)
	}
	refunds, _ := refundRepo.ListByExpenseID(ctx, expense.ID)
	if len(refunds) != 1 {
		t.Fatalf("refund rows = %d, want 1", len(refunds))
	}

	// 凭据号重复：只入账第一次，第二次失败且余额不变。
	if _, err := refundSvc.Register(ctx, actor, expense.ID, dto.CreateRefundRequest{
		Amount: 100, ReceivedDate: "2026-08-17", VoucherNo: "INT-0001",
	}); !errors.Is(err, ErrDuplicateVoucher) {
		t.Fatalf("error = %v, want ErrDuplicateVoucher", err)
	}
	gotExpense, _ = expenseRepo.FindByID(ctx, expense.ID)
	if gotExpense.RefundedAmount != 400 {
		t.Fatalf("expense refunded = %v, want 400 after duplicate", gotExpense.RefundedAmount)
	}

	// 退满剩余 600 后，剩余可退金额为 0。
	if _, err := refundSvc.Register(ctx, actor, expense.ID, dto.CreateRefundRequest{
		Amount: 600, ReceivedDate: "2026-08-20", Reason: "尾款减项", VoucherNo: "INT-0003",
	}); err != nil {
		t.Fatalf("final refund: %v", err)
	}
	detail, err := refundSvc.GetExpenseDetail(ctx, expense.ID)
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if detail.Refund.RefundedAmount != 1000 || detail.Refund.RefundableLeft != 0 || len(detail.Refund.Refunds) != 2 {
		t.Fatalf("detail = %+v", detail.Refund)
	}
	gotItem, _ = itemRepo.FindByID(ctx, item.ID)
	if gotItem.SpentAmount != 0 || gotItem.VarianceAmount != -5000 {
		t.Fatalf("item spent=%v variance=%v, want 0/-5000", gotItem.SpentAmount, gotItem.VarianceAmount)
	}
}
