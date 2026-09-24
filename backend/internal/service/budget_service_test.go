package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestBudgetServiceCreateAndGet(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	svc := NewBudgetService(newFakeBudgetRepo(), newFakeItemRepo(), audit, nil, testLogger())

	sheet, err := svc.Create(ctx, model.Actor{UserID: 1, Username: "admin"}, dto.CreateBudgetRequest{
		ProjectID:   "p-1",
		Name:        "整屋装修",
		TotalAmount: 20000,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if sheet.Status != constants.BudgetStatusDraft {
		t.Fatalf("status = %q, want Draft", sheet.Status)
	}
	if sheet.AvailableAmount != 20000 {
		t.Fatalf("available = %v, want 20000", sheet.AvailableAmount)
	}

	got, err := svc.Get(ctx, sheet.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "整屋装修" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestBudgetServiceDeleteNonDraft(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	svc := NewBudgetService(newFakeBudgetRepo(), newFakeItemRepo(), audit, nil, testLogger())

	sheet, err := svc.Create(ctx, model.Actor{UserID: 1}, dto.CreateBudgetRequest{ProjectID: "p-1", Name: "预算", TotalAmount: 1000, Status: constants.BudgetStatusActive})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.Delete(ctx, model.Actor{UserID: 1}, sheet.ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("error = %v, want ErrInvalidState", err)
	}
}

func TestBudgetServiceAdjust(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	svc := NewBudgetService(newFakeBudgetRepo(), newFakeItemRepo(), audit, nil, testLogger())

	sheet, err := svc.Create(ctx, model.Actor{UserID: 1}, dto.CreateBudgetRequest{ProjectID: "p-1", Name: "预算", TotalAmount: 1000})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	adjusted, err := svc.Adjust(ctx, model.Actor{UserID: 1}, sheet.ID, dto.AdjustBudgetRequest{TotalAmount: 1500, Reason: "增加主材预算"})
	if err != nil {
		t.Fatalf("adjust: %v", err)
	}
	if adjusted.TotalAmount != 1500 || adjusted.Version != 2 {
		t.Fatalf("unexpected adjusted: total=%v version=%d", adjusted.TotalAmount, adjusted.Version)
	}
}
