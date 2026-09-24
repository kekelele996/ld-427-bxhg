package service

import (
	"context"
	"errors"
	"testing"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestSupplierServiceCreateAndGet(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	svc := NewSupplierService(newFakeSupplierRepo(), audit, testLogger())

	supplier, err := svc.Create(ctx, model.Actor{UserID: 1}, dto.CreateSupplierRequest{
		Name:     "红星建材",
		Category: constants.SupplierCategoryMaterial,
		Rating:   5,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if supplier.Status != constants.SupplierStatusActive {
		t.Fatalf("status = %q, want Active", supplier.Status)
	}
	got, err := svc.Get(ctx, supplier.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "红星建材" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestSupplierServiceNotFound(t *testing.T) {
	svc := NewSupplierService(newFakeSupplierRepo(), NewAuditService(newFakeAuditRepo(), testLogger()), testLogger())
	if _, err := svc.Get(context.Background(), 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
