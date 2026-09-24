package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestRefundRepositoryCreateAndQuery(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	budgetRepo := NewBudgetRepository(db)
	itemRepo := NewItemRepository(db)
	expenseRepo := NewExpenseRepository(db)
	refundRepo := NewRefundRepository(db)

	sheet := &model.BudgetSheet{ProjectID: "p-1", Name: "预算", TotalAmount: 1000, AvailableAmount: 1000, Status: constants.BudgetStatusDraft, CreatedByID: 1, Version: 1}
	if err := budgetRepo.Create(ctx, sheet); err != nil {
		t.Fatalf("create sheet: %v", err)
	}
	item := &model.BudgetItem{BudgetSheetID: sheet.ID, Category: constants.BudgetCategoryLabor, BudgetAmount: 500, VarianceAmount: -500}
	if err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	expense := &model.ExpenseRecord{
		BudgetItemID: item.ID, Amount: 500, ExpenseDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PaymentMethod: constants.PaymentMethodBankTransfer, Status: constants.ExpenseStatusPaid, ApplicantID: 2,
	}
	if err := expenseRepo.Create(ctx, expense); err != nil {
		t.Fatalf("create expense: %v", err)
	}

	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	r1 := &model.ExpenseRefund{ExpenseID: expense.ID, Amount: 200, ReceivedDate: day, Reason: "退料", VoucherNo: "RV-001", RegisteredByID: 3}
	if err := refundRepo.Create(ctx, r1); err != nil {
		t.Fatalf("create refund 1: %v", err)
	}
	r2 := &model.ExpenseRefund{ExpenseID: expense.ID, Amount: 100.5, ReceivedDate: day.AddDate(0, 0, 1), Reason: "减项", VoucherNo: "RV-002", RegisteredByID: 3}
	if err := refundRepo.Create(ctx, r2); err != nil {
		t.Fatalf("create refund 2: %v", err)
	}

	// 凭据号重复：唯一约束兜底，返回哨兵错误。
	dup := &model.ExpenseRefund{ExpenseID: expense.ID, Amount: 50, ReceivedDate: day, Reason: "x", VoucherNo: "RV-001", RegisteredByID: 3}
	if err := refundRepo.Create(ctx, dup); !errors.Is(err, ErrDuplicateRefundVoucher) {
		t.Fatalf("duplicate create error = %v, want ErrDuplicateRefundVoucher", err)
	}

	exists, err := refundRepo.ExistsByVoucherNo(ctx, "RV-001")
	if err != nil || !exists {
		t.Fatalf("exists RV-001 = %v, err = %v", exists, err)
	}
	exists, err = refundRepo.ExistsByVoucherNo(ctx, "MISSING")
	if err != nil || exists {
		t.Fatalf("exists MISSING = %v, err = %v", exists, err)
	}

	total, err := refundRepo.SumAmountByExpenseID(ctx, expense.ID)
	if err != nil {
		t.Fatalf("sum: %v", err)
	}
	if total != 300.5 {
		t.Fatalf("sum = %v, want 300.5", total)
	}
	if total, _ := refundRepo.SumAmountByExpenseID(ctx, 9999); total != 0 {
		t.Fatalf("sum for missing expense = %v, want 0", total)
	}

	refunds, err := refundRepo.ListByExpenseID(ctx, expense.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(refunds) != 2 {
		t.Fatalf("len = %d, want 2", len(refunds))
	}
	// 到账日期倒序。
	if refunds[0].VoucherNo != "RV-002" || refunds[1].VoucherNo != "RV-001" {
		t.Fatalf("order = %s, %s", refunds[0].VoucherNo, refunds[1].VoucherNo)
	}

	got, err := refundRepo.FindByID(ctx, r1.ID)
	if err != nil || got.VoucherNo != "RV-001" {
		t.Fatalf("find by id = %+v, err = %v", got, err)
	}
	if _, err := refundRepo.FindByID(ctx, 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("find missing error = %v, want ErrNotFound", err)
	}
}
