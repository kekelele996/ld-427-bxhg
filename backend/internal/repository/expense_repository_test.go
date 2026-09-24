package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestExpenseRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	budgetRepo := NewBudgetRepository(db)
	itemRepo := NewItemRepository(db)
	expenseRepo := NewExpenseRepository(db)

	sheet := &model.BudgetSheet{ProjectID: "p-1", Name: "预算", TotalAmount: 1000, AvailableAmount: 1000, Status: constants.BudgetStatusDraft, CreatedByID: 1, Version: 1}
	if err := budgetRepo.Create(ctx, sheet); err != nil {
		t.Fatalf("create sheet: %v", err)
	}
	item := &model.BudgetItem{BudgetSheetID: sheet.ID, Category: constants.BudgetCategoryLabor, BudgetAmount: 500, VarianceAmount: -500}
	if err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("create item: %v", err)
	}

	record := &model.ExpenseRecord{
		BudgetItemID:  item.ID,
		Amount:        150,
		ExpenseDate:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PaymentMethod: constants.PaymentMethodBankTransfer,
		Status:        constants.ExpenseStatusDraft,
		ApplicantID:   2,
	}
	if err := expenseRepo.Create(ctx, record); err != nil {
		t.Fatalf("create expense: %v", err)
	}

	records, total, err := expenseRepo.List(ctx, ExpenseListFilter{BudgetItemID: item.ID, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list expenses: %v", err)
	}
	if total != 1 || len(records) != 1 {
		t.Fatalf("total=%d len=%d", total, len(records))
	}

	record.Status = constants.ExpenseStatusSubmitted
	if err := expenseRepo.Update(ctx, record); err != nil {
		t.Fatalf("update expense: %v", err)
	}
	got, err := expenseRepo.FindByID(ctx, record.ID)
	if err != nil {
		t.Fatalf("find expense: %v", err)
	}
	if got.Status != constants.ExpenseStatusSubmitted {
		t.Fatalf("status = %q", got.Status)
	}

	if _, err := expenseRepo.FindByID(ctx, 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
