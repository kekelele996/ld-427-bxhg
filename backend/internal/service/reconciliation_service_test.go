package service

import (
	"context"
	"errors"
	"testing"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestReconciliationServiceFlow(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	svc := NewReconciliationService(newFakeReconciliationRepo(), audit, testLogger())

	record, err := svc.Create(ctx, model.Actor{UserID: 1}, dto.CreateReconciliationRequest{
		ProjectID:     "p-1",
		Period:        "2026-08",
		SupplierID:    1,
		PayableAmount: 1000,
		PaidAmount:    300,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if record.UnpaidAmount != 700 {
		t.Fatalf("unpaid = %v, want 700", record.UnpaidAmount)
	}

	confirmed, err := svc.Confirm(ctx, model.Actor{UserID: 1}, record.ID)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirmed.Status != constants.ReconciliationStatusConfirmed {
		t.Fatalf("status = %q, want Confirmed", confirmed.Status)
	}

	if _, err := svc.Confirm(ctx, model.Actor{UserID: 1}, record.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("error = %v, want ErrInvalidState", err)
	}
}
