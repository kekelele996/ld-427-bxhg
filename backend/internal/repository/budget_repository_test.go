package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestBudgetRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewBudgetRepository(newTestDB(t))

	sheet := &model.BudgetSheet{
		ProjectID:       "p-100",
		Name:            "客厅装修",
		TotalAmount:     10000,
		AvailableAmount: 10000,
		Status:          constants.BudgetStatusDraft,
		CreatedByID:     1,
		Version:         1,
	}
	if err := repo.Create(ctx, sheet); err != nil {
		t.Fatalf("create: %v", err)
	}
	if sheet.ID == 0 {
		t.Fatalf("expected id to be assigned")
	}

	got, err := repo.FindByID(ctx, sheet.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got.Name != sheet.Name {
		t.Fatalf("name = %q, want %q", got.Name, sheet.Name)
	}

	sheets, total, err := repo.List(ctx, BudgetListFilter{ProjectID: "p-100", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(sheets) != 1 {
		t.Fatalf("list total=%d len=%d", total, len(sheets))
	}

	sheet.Status = constants.BudgetStatusActive
	sheet.TotalAmount = 12000
	sheet.AvailableAmount = 12000
	if err := repo.Update(ctx, sheet); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err = repo.FindByID(ctx, sheet.ID)
	if err != nil {
		t.Fatalf("find after update: %v", err)
	}
	if got.Status != constants.BudgetStatusActive || got.TotalAmount != 12000 {
		t.Fatalf("unexpected updated value: %+v", got)
	}

	if err := repo.Delete(ctx, sheet.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, sheet.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete error = %v, want ErrNotFound", err)
	}
}

func TestBudgetRepositoryFindByIDNotFound(t *testing.T) {
	repo := NewBudgetRepository(newTestDB(t))
	if _, err := repo.FindByID(context.Background(), 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
