package service

import (
	"context"
	"errors"
	"testing"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestExpenseServiceFullFlow(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	budgetRepo := newFakeBudgetRepo()
	itemRepo := newFakeItemRepo()
	budgetSvc := NewBudgetService(budgetRepo, itemRepo, audit, nil, testLogger())
	expenseSvc := NewExpenseService(newFakeExpenseRepo(), itemRepo, budgetRepo, audit, nil, testLogger())

	sheet, err := budgetSvc.Create(ctx, model.Actor{UserID: 1, Username: "project"}, dto.CreateBudgetRequest{ProjectID: "p-1", Name: "整屋装修", TotalAmount: 10000})
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	item, err := NewItemService(itemRepo, budgetRepo, audit, nil, testLogger()).Create(ctx, model.Actor{UserID: 1}, sheet.ID, dto.CreateItemRequest{Category: constants.BudgetCategoryMaterial, BudgetAmount: 5000})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	record, err := expenseSvc.Create(ctx, model.Actor{UserID: 2, Username: "accountant"}, dto.CreateExpenseRequest{
		BudgetItemID:  item.ID,
		Amount:        500,
		ExpenseDate:   "2026-08-01",
		PaymentMethod: constants.PaymentMethodBankTransfer,
	})
	if err != nil {
		t.Fatalf("create expense: %v", err)
	}
	if record.Status != constants.ExpenseStatusDraft {
		t.Fatalf("status = %q, want Draft", record.Status)
	}

	if _, err := expenseSvc.Submit(ctx, model.Actor{UserID: 2}, record.ID); err != nil {
		t.Fatalf("submit: %v", err)
	}
	budget, _ := budgetRepo.FindByID(ctx, sheet.ID)
	if budget.FrozenAmount != 500 || budget.AvailableAmount != 9500 {
		t.Fatalf("after submit frozen=%v available=%v", budget.FrozenAmount, budget.AvailableAmount)
	}

	if _, err := expenseSvc.Approve(ctx, model.Actor{UserID: 3, Username: "finance"}, record.ID, dto.ApproveExpenseRequest{ApprovalComment: "同意"}); err != nil {
		t.Fatalf("approve: %v", err)
	}
	budget, _ = budgetRepo.FindByID(ctx, sheet.ID)
	updatedItem, _ := itemRepo.FindByID(ctx, item.ID)
	if budget.FrozenAmount != 0 || budget.SpentAmount != 500 || budget.AvailableAmount != 9500 {
		t.Fatalf("after approve frozen=%v spent=%v available=%v", budget.FrozenAmount, budget.SpentAmount, budget.AvailableAmount)
	}
	if updatedItem.SpentAmount != 500 || updatedItem.VarianceAmount != -4500 {
		t.Fatalf("item spent=%v variance=%v", updatedItem.SpentAmount, updatedItem.VarianceAmount)
	}

	paid, err := expenseSvc.Pay(ctx, model.Actor{UserID: 3}, record.ID, dto.PayExpenseRequest{PaymentDate: "2026-08-10"})
	if err != nil {
		t.Fatalf("pay: %v", err)
	}
	if paid.Status != constants.ExpenseStatusPaid {
		t.Fatalf("status = %q, want Paid", paid.Status)
	}
}

func TestExpenseServiceSubmitInsufficient(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	budgetRepo := newFakeBudgetRepo()
	itemRepo := newFakeItemRepo()
	expenseSvc := NewExpenseService(newFakeExpenseRepo(), itemRepo, budgetRepo, audit, nil, testLogger())

	sheet := &model.BudgetSheet{ProjectID: "p-1", Name: "预算", TotalAmount: 100, AvailableAmount: 100, Status: constants.BudgetStatusDraft, CreatedByID: 1, Version: 1}
	if err := budgetRepo.Create(ctx, sheet); err != nil {
		t.Fatalf("create sheet: %v", err)
	}
	item := &model.BudgetItem{BudgetSheetID: sheet.ID, Category: constants.BudgetCategoryLabor, BudgetAmount: 100, VarianceAmount: -100}
	if err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	record, err := expenseSvc.Create(ctx, model.Actor{UserID: 1}, dto.CreateExpenseRequest{BudgetItemID: item.ID, Amount: 200, ExpenseDate: "2026-08-01", PaymentMethod: constants.PaymentMethodCash})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := expenseSvc.Submit(ctx, model.Actor{UserID: 1}, record.ID); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("error = %v, want ErrInsufficientBalance", err)
	}
}
