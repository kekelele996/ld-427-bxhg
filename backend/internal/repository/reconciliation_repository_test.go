package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestReconciliationRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewReconciliationRepository(newTestDB(t))

	record := &model.Reconciliation{
		ProjectID:     "p-1",
		Period:        "2026-08",
		SupplierID:    9,
		PayableAmount: 1000,
		PaidAmount:    300,
		UnpaidAmount:  700,
		Status:        constants.ReconciliationStatusPending,
	}
	if err := repo.Create(ctx, record); err != nil {
		t.Fatalf("create reconciliation: %v", err)
	}

	got, err := repo.FindByID(ctx, record.ID)
	if err != nil {
		t.Fatalf("find reconciliation: %v", err)
	}
	if got.Period != record.Period {
		t.Fatalf("period = %q", got.Period)
	}

	records, total, err := repo.List(ctx, ReconciliationListFilter{ProjectID: "p-1", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list reconciliations: %v", err)
	}
	if total != 1 || len(records) != 1 {
		t.Fatalf("total=%d len=%d", total, len(records))
	}

	record.Status = constants.ReconciliationStatusConfirmed
	if err := repo.Update(ctx, record); err != nil {
		t.Fatalf("update reconciliation: %v", err)
	}

	if err := repo.Delete(ctx, record.ID); err != nil {
		t.Fatalf("delete reconciliation: %v", err)
	}
	if _, err := repo.FindByID(ctx, record.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
