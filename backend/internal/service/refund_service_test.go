package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func parseDate(t *testing.T, value string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse date %q: %v", value, err)
	}
	return d
}

func setupPaidExpense(t *testing.T, ctx context.Context, audit *AuditService) (*fakeBudgetRepo, *fakeItemRepo, *fakeExpenseRepo, uint, uint, uint) {
	t.Helper()
	budgetRepo := newFakeBudgetRepo()
	itemRepo := newFakeItemRepo()
	expenseRepo := newFakeExpenseRepo()
	budgetSvc := NewBudgetService(budgetRepo, itemRepo, audit, nil, testLogger())
	itemSvc := NewItemService(itemRepo, budgetRepo, audit, nil, testLogger())
	expenseSvc := NewExpenseService(expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger())

	sheet, err := budgetSvc.Create(ctx, model.Actor{UserID: 1, Username: "project"}, dto.CreateBudgetRequest{ProjectID: "p-1", Name: "整屋装修", TotalAmount: 10000})
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	item, err := itemSvc.Create(ctx, model.Actor{UserID: 1}, sheet.ID, dto.CreateItemRequest{Category: constants.BudgetCategoryMaterial, BudgetAmount: 5000})
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
	if _, err := expenseSvc.Submit(ctx, model.Actor{UserID: 2}, record.ID); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := expenseSvc.Approve(ctx, model.Actor{UserID: 3, Username: "finance"}, record.ID, dto.ApproveExpenseRequest{ApprovalComment: "同意"}); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if _, err := expenseSvc.Pay(ctx, model.Actor{UserID: 3}, record.ID, dto.PayExpenseRequest{PaymentDate: "2026-08-10"}); err != nil {
		t.Fatalf("pay: %v", err)
	}
	return budgetRepo, itemRepo, expenseRepo, sheet.ID, item.ID, record.ID
}

func TestRefundServiceRegisterReversesBudget(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	budgetRepo, itemRepo, expenseRepo, sheetID, itemID, expenseID := setupPaidExpense(t, ctx, audit)
	refundRepo := newFakeRefundRepo()
	refundSvc := NewRefundService(refundRepo, expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger())

	r1, err := refundSvc.Register(ctx, model.Actor{UserID: 3, Username: "finance"}, expenseID, dto.RegisterRefundRequest{
		Amount:       200,
		ReceivedDate: "2026-09-01",
		Reason:       "瓷砖退料",
		VoucherNo:    "V-20260901-001",
	})
	if err != nil {
		t.Fatalf("register first refund: %v", err)
	}
	if r1.Amount != 200 || r1.RegisteredByID != 3 {
		t.Fatalf("refund = %+v", r1)
	}

	budget, _ := budgetRepo.FindByID(ctx, sheetID)
	item, _ := itemRepo.FindByID(ctx, itemID)
	if budget.SpentAmount != 300 || budget.FrozenAmount != 0 || budget.AvailableAmount != 9700 {
		t.Fatalf("after first refund budget spent=%v frozen=%v available=%v", budget.SpentAmount, budget.FrozenAmount, budget.AvailableAmount)
	}
	if item.SpentAmount != 300 || item.VarianceAmount != -4700 {
		t.Fatalf("after first refund item spent=%v variance=%v", item.SpentAmount, item.VarianceAmount)
	}

	// 分次退款，累计等于实付金额时允许。
	if _, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, expenseID, dto.RegisterRefundRequest{
		Amount:       300,
		ReceivedDate: "2026-09-05",
		Reason:       "减项",
		VoucherNo:    "V-20260905-002",
	}); err != nil {
		t.Fatalf("register second refund: %v", err)
	}
	budget, _ = budgetRepo.FindByID(ctx, sheetID)
	item, _ = itemRepo.FindByID(ctx, itemID)
	if budget.SpentAmount != 0 || budget.AvailableAmount != 10000 {
		t.Fatalf("after full refund budget spent=%v available=%v", budget.SpentAmount, budget.AvailableAmount)
	}
	if item.SpentAmount != 0 || item.VarianceAmount != -5000 {
		t.Fatalf("after full refund item spent=%v variance=%v", item.SpentAmount, item.VarianceAmount)
	}

	summary, err := refundSvc.GetSummary(ctx, expenseID)
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if summary.RefundedAmount != 500 || summary.RefundableLeft != 0 || len(summary.Records) != 2 {
		t.Fatalf("summary = %+v", summary)
	}

	detail, err := refundSvc.GetExpenseDetail(ctx, expenseID)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if detail.Status != constants.ExpenseStatusPaid || detail.RefundedAmount != 500 || detail.RefundableLeft != 0 || len(detail.Refunds) != 2 {
		t.Fatalf("detail = %+v", detail)
	}
}

func TestRefundServiceExceedsPaidFailsAndBalancesUnchanged(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	budgetRepo, itemRepo, expenseRepo, sheetID, itemID, expenseID := setupPaidExpense(t, ctx, audit)
	refundRepo := newFakeRefundRepo()
	refundSvc := NewRefundService(refundRepo, expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger())

	if _, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, expenseID, dto.RegisterRefundRequest{
		Amount: 300, ReceivedDate: "2026-09-01", Reason: "退料", VoucherNo: "V-1",
	}); err != nil {
		t.Fatalf("register refund: %v", err)
	}

	// 累计 300 + 300 > 实付 500，操作失败。
	_, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, expenseID, dto.RegisterRefundRequest{
		Amount: 300, ReceivedDate: "2026-09-02", Reason: "退料", VoucherNo: "V-2",
	})
	if !errors.Is(err, ErrRefundExceedsPaid) {
		t.Fatalf("error = %v, want ErrRefundExceedsPaid", err)
	}

	// 操作失败：支出、分项和预算余额保持不变，退款不入账。
	budget, _ := budgetRepo.FindByID(ctx, sheetID)
	item, _ := itemRepo.FindByID(ctx, itemID)
	if budget.SpentAmount != 200 || budget.AvailableAmount != 9800 {
		t.Fatalf("budget changed after failed refund: spent=%v available=%v", budget.SpentAmount, budget.AvailableAmount)
	}
	if item.SpentAmount != 200 {
		t.Fatalf("item spent changed after failed refund: %v", item.SpentAmount)
	}
	if n, _ := refundRepo.SumAmountByExpenseID(ctx, expenseID); n != 300 {
		t.Fatalf("failed refund was recorded, total = %v", n)
	}
}

func TestRefundServiceDuplicateVoucherOnlyFirstPosted(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	budgetRepo, itemRepo, expenseRepo, _, _, expenseID := setupPaidExpense(t, ctx, audit)
	refundRepo := newFakeRefundRepo()
	refundSvc := NewRefundService(refundRepo, expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger())

	req := dto.RegisterRefundRequest{Amount: 100, ReceivedDate: "2026-09-01", Reason: "退料", VoucherNo: "DUP-1"}
	if _, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, expenseID, req); err != nil {
		t.Fatalf("first register: %v", err)
	}
	// 凭据号重复，第二次失败。
	if _, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, expenseID, req); !errors.Is(err, ErrDuplicateVoucherNo) {
		t.Fatalf("error = %v, want ErrDuplicateVoucherNo", err)
	}
	refunds, _ := refundRepo.ListByExpenseID(ctx, expenseID)
	if len(refunds) != 1 || refunds[0].VoucherNo != "DUP-1" {
		t.Fatalf("refunds = %+v, only first posting expected", refunds)
	}
}

func TestRefundServiceRejectsNonPaidExpense(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	budgetRepo := newFakeBudgetRepo()
	itemRepo := newFakeItemRepo()
	expenseRepo := newFakeExpenseRepo()
	refundSvc := NewRefundService(newFakeRefundRepo(), expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger())

	sheet := &model.BudgetSheet{ProjectID: "p-1", Name: "预算", TotalAmount: 1000, AvailableAmount: 1000, Status: constants.BudgetStatusDraft, CreatedByID: 1, Version: 1}
	if err := budgetRepo.Create(ctx, sheet); err != nil {
		t.Fatalf("create sheet: %v", err)
	}
	item := &model.BudgetItem{BudgetSheetID: sheet.ID, Category: constants.BudgetCategoryLabor, BudgetAmount: 500, VarianceAmount: -500}
	if err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	record := &model.ExpenseRecord{
		BudgetItemID: item.ID, Amount: 100, ExpenseDate: parseDate(t, "2026-08-01"),
		PaymentMethod: constants.PaymentMethodCash, Status: constants.ExpenseStatusApproved, ApplicantID: 2,
	}
	if err := expenseRepo.Create(ctx, record); err != nil {
		t.Fatalf("create expense: %v", err)
	}

	if _, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, record.ID, dto.RegisterRefundRequest{
		Amount: 10, ReceivedDate: "2026-09-01", Reason: "退料", VoucherNo: "V-9",
	}); !errors.Is(err, ErrRefundNotAllowed) {
		t.Fatalf("error = %v, want ErrRefundNotAllowed", err)
	}
}

func TestRefundServiceExpenseNotFound(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	refundSvc := NewRefundService(newFakeRefundRepo(), newFakeExpenseRepo(), newFakeItemRepo(), newFakeBudgetRepo(), audit, nil, testLogger())

	if _, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, 999, dto.RegisterRefundRequest{
		Amount: 10, ReceivedDate: "2026-09-01", Reason: "退料", VoucherNo: "V-9",
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("register error = %v, want ErrNotFound", err)
	}
	if _, err := refundSvc.GetSummary(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("summary error = %v, want ErrNotFound", err)
	}
}
