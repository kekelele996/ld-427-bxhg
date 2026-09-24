package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestRefundRepositoryCreateAndList(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	budgetRepo := NewBudgetRepository(db)
	itemRepo := NewItemRepository(db)
	expenseRepo := NewExpenseRepository(db)
	refundRepo := NewRefundRepository(db)

	sheet := &model.BudgetSheet{ProjectID: "p-1", Name: "预算", TotalAmount: 1000, AvailableAmount: 1000, Status: constants.BudgetStatusActive, CreatedByID: 1, Version: 1}
	if err := budgetRepo.Create(ctx, sheet); err != nil {
		t.Fatalf("create sheet: %v", err)
	}
	item := &model.BudgetItem{BudgetSheetID: sheet.ID, Category: constants.BudgetCategoryMaterial, BudgetAmount: 1000, VarianceAmount: -1000}
	if err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	expense := &model.ExpenseRecord{
		BudgetItemID:  item.ID,
		Amount:        500,
		ExpenseDate:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PaymentMethod: constants.PaymentMethodBankTransfer,
		Status:        constants.ExpenseStatusPaid,
		ApplicantID:   2,
	}
	if err := expenseRepo.Create(ctx, expense); err != nil {
		t.Fatalf("create expense: %v", err)
	}

	first := &model.RefundRecord{
		ExpenseID:      expense.ID,
		Amount:         200,
		ReceivedDate:   time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
		Reason:         "退料",
		VoucherNo:      "V-0001",
		RegisteredByID: 3,
	}
	if err := refundRepo.Create(ctx, first); err != nil {
		t.Fatalf("create first refund: %v", err)
	}

	// 相同凭据号再次入账必须失败，且只有第一次记录保留。
	dup := &model.RefundRecord{
		ExpenseID:      expense.ID,
		Amount:         100,
		ReceivedDate:   time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC),
		VoucherNo:      "V-0001",
		RegisteredByID: 3,
	}
	if err := refundRepo.Create(ctx, dup); !errors.Is(err, ErrDuplicateVoucherNo) {
		t.Fatalf("error = %v, want ErrDuplicateVoucherNo", err)
	}

	second := &model.RefundRecord{
		ExpenseID:      expense.ID,
		Amount:         50,
		ReceivedDate:   time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC),
		Reason:         "减项",
		VoucherNo:      "V-0002",
		RegisteredByID: 3,
	}
	if err := refundRepo.Create(ctx, second); err != nil {
		t.Fatalf("create second refund: %v", err)
	}

	refunds, err := refundRepo.ListByExpenseID(ctx, expense.ID)
	if err != nil {
		t.Fatalf("list refunds: %v", err)
	}
	if len(refunds) != 2 {
		t.Fatalf("len = %d, want 2", len(refunds))
	}
	if refunds[0].VoucherNo != "V-0001" || refunds[1].VoucherNo != "V-0002" {
		t.Fatalf("refunds not ordered by id asc: %s, %s", refunds[0].VoucherNo, refunds[1].VoucherNo)
	}

	if _, err := refundRepo.FindByID(ctx, first.ID); err != nil {
		t.Fatalf("find refund: %v", err)
	}
	if _, err := refundRepo.FindByID(ctx, 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestExpenseRepositoryFindForUpdate(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	budgetRepo := NewBudgetRepository(db)
	itemRepo := NewItemRepository(db)
	expenseRepo := NewExpenseRepository(db)

	sheet := &model.BudgetSheet{ProjectID: "p-1", Name: "预算", TotalAmount: 100, AvailableAmount: 100, Status: constants.BudgetStatusDraft, CreatedByID: 1, Version: 1}
	if err := budgetRepo.Create(ctx, sheet); err != nil {
		t.Fatalf("create sheet: %v", err)
	}
	item := &model.BudgetItem{BudgetSheetID: sheet.ID, Category: constants.BudgetCategoryLabor, BudgetAmount: 100}
	if err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	expense := &model.ExpenseRecord{
		BudgetItemID: item.ID, Amount: 80,
		ExpenseDate:   time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		PaymentMethod: constants.PaymentMethodCash,
		Status:        constants.ExpenseStatusPaid,
		ApplicantID:   2,
	}
	if err := expenseRepo.Create(ctx, expense); err != nil {
		t.Fatalf("create expense: %v", err)
	}
	got, err := expenseRepo.FindByIDForUpdate(ctx, expense.ID)
	if err != nil {
		t.Fatalf("find for update: %v", err)
	}
	if got.ID != expense.ID {
		t.Fatalf("id = %d, want %d", got.ID, expense.ID)
	}
	if _, err := expenseRepo.FindByIDForUpdate(ctx, 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
