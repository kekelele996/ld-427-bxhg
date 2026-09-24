package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestItemRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	budgetRepo := NewBudgetRepository(db)
	itemRepo := NewItemRepository(db)

	sheet := &model.BudgetSheet{ProjectID: "p-1", Name: "预算", TotalAmount: 1000, AvailableAmount: 1000, Status: constants.BudgetStatusDraft, CreatedByID: 1, Version: 1}
	if err := budgetRepo.Create(ctx, sheet); err != nil {
		t.Fatalf("create sheet: %v", err)
	}

	item := &model.BudgetItem{
		BudgetSheetID:  sheet.ID,
		Category:       constants.BudgetCategoryMaterial,
		SubCategory:    "瓷砖",
		BudgetAmount:   300,
		VarianceAmount: -300,
	}
	if err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected item id")
	}

	items, err := itemRepo.ListByBudgetID(ctx, sheet.ID)
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}

	item.SpentAmount = 100
	item.VarianceAmount = -200
	if err := itemRepo.Update(ctx, item); err != nil {
		t.Fatalf("update item: %v", err)
	}
	got, err := itemRepo.FindByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("find item: %v", err)
	}
	if got.SpentAmount != 100 {
		t.Fatalf("spent = %v, want 100", got.SpentAmount)
	}

	if err := itemRepo.Delete(ctx, item.ID); err != nil {
		t.Fatalf("delete item: %v", err)
	}
	if _, err := itemRepo.FindByID(ctx, item.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
