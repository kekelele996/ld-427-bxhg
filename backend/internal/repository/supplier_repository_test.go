package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func TestSupplierRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewSupplierRepository(newTestDB(t))

	supplier := &model.Supplier{
		Name:     "红星建材",
		Category: constants.SupplierCategoryMaterial,
		Contact:  "张三",
		Phone:    "13800000000",
		Status:   constants.SupplierStatusActive,
		Rating:   5,
	}
	if err := repo.Create(ctx, supplier); err != nil {
		t.Fatalf("create supplier: %v", err)
	}

	got, err := repo.FindByID(ctx, supplier.ID)
	if err != nil {
		t.Fatalf("find supplier: %v", err)
	}
	if got.Name != supplier.Name {
		t.Fatalf("name = %q", got.Name)
	}

	suppliers, total, err := repo.List(ctx, SupplierListFilter{Status: string(constants.SupplierStatusActive), Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list suppliers: %v", err)
	}
	if total != 1 || len(suppliers) != 1 {
		t.Fatalf("total=%d len=%d", total, len(suppliers))
	}

	supplier.Status = constants.SupplierStatusSuspended
	if err := repo.Update(ctx, supplier); err != nil {
		t.Fatalf("update supplier: %v", err)
	}

	if err := repo.Delete(ctx, supplier.ID); err != nil {
		t.Fatalf("delete supplier: %v", err)
	}
	if _, err := repo.FindByID(ctx, supplier.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
