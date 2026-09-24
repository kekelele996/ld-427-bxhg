package service

import (
	"context"
	"errors"
	"testing"

	"github.com/renovation/renovation-budget-api/internal/constants"
	"github.com/renovation/renovation-budget-api/internal/dto"
	"github.com/renovation/renovation-budget-api/internal/model"
)

func setupRefundScenario(t *testing.T) (*RefundService, repositoryShim, *model.ExpenseRecord, *model.BudgetSheet, *model.BudgetItem) {
	t.Helper()
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	budgetRepo := newFakeBudgetRepo()
	itemRepo := newFakeItemRepo()
	expenseRepo := newFakeExpenseRepo()
	refundRepo := newFakeRefundRepo()
	budgetSvc := NewBudgetService(budgetRepo, itemRepo, audit, nil, testLogger())
	expenseSvc := NewExpenseService(expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger())
	refundSvc := NewRefundService(newFakeTxManager(), refundRepo, expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger())

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
		Amount:        1000,
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
	return refundSvc, repositoryShim{budgetRepo, itemRepo, expenseRepo, refundRepo}, record, sheet, item
}

type repositoryShim struct {
	budget  *fakeBudgetRepo
	item    *fakeItemRepo
	expense *fakeExpenseRepo
	refund  *fakeRefundRepo
}

func TestRefundServiceReversesBudget(t *testing.T) {
	ctx := context.Background()
	refundSvc, repos, record, sheet, item := setupRefundScenario(t)

	refund, err := refundSvc.Register(ctx, model.Actor{UserID: 3, Username: "finance"}, record.ID, dto.CreateRefundRequest{
		Amount:       300,
		ReceivedDate: "2026-08-15",
		Reason:       "瓷砖退料",
		VoucherNo:    "RF-0001",
	})
	if err != nil {
		t.Fatalf("register refund: %v", err)
	}
	if refund.ID == 0 || refund.VoucherNo != "RF-0001" {
		t.Fatalf("unexpected refund: %+v", refund)
	}

	expense, _ := repos.expense.FindByID(ctx, record.ID)
	if expense.RefundedAmount != 300 {
		t.Fatalf("expense refunded = %v, want 300", expense.RefundedAmount)
	}
	updatedItem, _ := repos.item.FindByID(ctx, item.ID)
	if updatedItem.SpentAmount != 700 {
		t.Fatalf("item spent = %v, want 700", updatedItem.SpentAmount)
	}
	if updatedItem.VarianceAmount != -4300 {
		t.Fatalf("item variance = %v, want -4300", updatedItem.VarianceAmount)
	}
	updatedBudget, _ := repos.budget.FindByID(ctx, sheet.ID)
	if updatedBudget.SpentAmount != 700 || updatedBudget.AvailableAmount != 9300 {
		t.Fatalf("budget spent=%v available=%v, want 700/9300", updatedBudget.SpentAmount, updatedBudget.AvailableAmount)
	}

	// 第二笔分次退款，累计恰好等于实付金额。
	if _, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, record.ID, dto.CreateRefundRequest{
		Amount:       700,
		ReceivedDate: "2026-08-20",
		Reason:       "减项",
		VoucherNo:    "RF-0002",
	}); err != nil {
		t.Fatalf("register second refund: %v", err)
	}
	expense, _ = repos.expense.FindByID(ctx, record.ID)
	if expense.RefundedAmount != 1000 {
		t.Fatalf("expense refunded = %v, want 1000", expense.RefundedAmount)
	}

	detail, err := refundSvc.GetExpenseDetail(ctx, record.ID)
	if err != nil {
		t.Fatalf("get detail: %v", err)
	}
	if detail.Refund.RefundedAmount != 1000 || detail.Refund.RefundableLeft != 0 {
		t.Fatalf("refunded=%v left=%v, want 1000/0", detail.Refund.RefundedAmount, detail.Refund.RefundableLeft)
	}
	if len(detail.Refund.Refunds) != 2 {
		t.Fatalf("refund records = %d, want 2", len(detail.Refund.Refunds))
	}
}

func TestRefundServiceExceedFailsAndLeavesBalancesUntouched(t *testing.T) {
	ctx := context.Background()
	refundSvc, repos, record, sheet, item := setupRefundScenario(t)

	if _, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, record.ID, dto.CreateRefundRequest{
		Amount:       600,
		ReceivedDate: "2026-08-15",
		Reason:       "退料",
		VoucherNo:    "RF-0001",
	}); err != nil {
		t.Fatalf("register first refund: %v", err)
	}

	_, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, record.ID, dto.CreateRefundRequest{
		Amount:       500,
		ReceivedDate: "2026-08-16",
		Reason:       "再次退料",
		VoucherNo:    "RF-0002",
	})
	if !errors.Is(err, ErrRefundExceeded) {
		t.Fatalf("error = %v, want ErrRefundExceeded", err)
	}

	// 失败后支出、分项、预算余额均保持不变，且不产生第二条退款记录。
	expense, _ := repos.expense.FindByID(ctx, record.ID)
	if expense.RefundedAmount != 600 {
		t.Fatalf("expense refunded = %v, want 600", expense.RefundedAmount)
	}
	updatedItem, _ := repos.item.FindByID(ctx, item.ID)
	if updatedItem.SpentAmount != 400 {
		t.Fatalf("item spent = %v, want 400", updatedItem.SpentAmount)
	}
	updatedBudget, _ := repos.budget.FindByID(ctx, sheet.ID)
	if updatedBudget.SpentAmount != 400 || updatedBudget.AvailableAmount != 9600 {
		t.Fatalf("budget spent=%v available=%v, want 400/9600", updatedBudget.SpentAmount, updatedBudget.AvailableAmount)
	}
	refunds, _ := repos.refund.ListByExpenseID(ctx, record.ID)
	if len(refunds) != 1 {
		t.Fatalf("refund records = %d, want 1", len(refunds))
	}
}

func TestRefundServiceDuplicateVoucherOnlyFirstPosted(t *testing.T) {
	ctx := context.Background()
	refundSvc, repos, record, sheet, item := setupRefundScenario(t)
	actor := model.Actor{UserID: 3, Username: "finance"}
	req := dto.CreateRefundRequest{Amount: 100, ReceivedDate: "2026-08-15", Reason: "退料", VoucherNo: "RF-DUP"}
	if _, err := refundSvc.Register(ctx, actor, record.ID, req); err != nil {
		t.Fatalf("register first refund: %v", err)
	}
	if _, err := refundSvc.Register(ctx, actor, record.ID, req); !errors.Is(err, ErrDuplicateVoucher) {
		t.Fatalf("error = %v, want ErrDuplicateVoucher", err)
	}

	expense, _ := repos.expense.FindByID(ctx, record.ID)
	if expense.RefundedAmount != 100 {
		t.Fatalf("expense refunded = %v, want 100", expense.RefundedAmount)
	}
	updatedItem, _ := repos.item.FindByID(ctx, item.ID)
	if updatedItem.SpentAmount != 900 {
		t.Fatalf("item spent = %v, want 900", updatedItem.SpentAmount)
	}
	updatedBudget, _ := repos.budget.FindByID(ctx, sheet.ID)
	if updatedBudget.SpentAmount != 900 {
		t.Fatalf("budget spent = %v, want 900", updatedBudget.SpentAmount)
	}
	refunds, _ := repos.refund.ListByExpenseID(ctx, record.ID)
	if len(refunds) != 1 {
		t.Fatalf("refund records = %d, want 1", len(refunds))
	}
}

func TestRefundServiceRejectsUnpaidExpense(t *testing.T) {
	ctx := context.Background()
	audit := NewAuditService(newFakeAuditRepo(), testLogger())
	budgetRepo := newFakeBudgetRepo()
	itemRepo := newFakeItemRepo()
	expenseRepo := newFakeExpenseRepo()
	refundSvc := NewRefundService(newFakeTxManager(), newFakeRefundRepo(), expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger())
	budgetSvc := NewBudgetService(budgetRepo, itemRepo, audit, nil, testLogger())

	sheet, err := budgetSvc.Create(ctx, model.Actor{UserID: 1}, dto.CreateBudgetRequest{ProjectID: "p-2", Name: "预算", TotalAmount: 1000})
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	item, err := NewItemService(itemRepo, budgetRepo, audit, nil, testLogger()).Create(ctx, model.Actor{UserID: 1}, sheet.ID, dto.CreateItemRequest{Category: constants.BudgetCategoryLabor, BudgetAmount: 1000})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	record, err := NewExpenseService(expenseRepo, itemRepo, budgetRepo, audit, nil, testLogger()).Create(ctx, model.Actor{UserID: 2}, dto.CreateExpenseRequest{
		BudgetItemID: item.ID, Amount: 200, ExpenseDate: "2026-09-01", PaymentMethod: constants.PaymentMethodCash,
	})
	if err != nil {
		t.Fatalf("create expense: %v", err)
	}
	// 仍处于 Draft 状态，登记退款应失败。
	if _, err := refundSvc.Register(ctx, model.Actor{UserID: 3}, record.ID, dto.CreateRefundRequest{
		Amount: 100, ReceivedDate: "2026-09-05", VoucherNo: "RF-X1",
	}); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("error = %v, want ErrInvalidState", err)
	}

	if _, err := refundSvc.GetExpenseDetail(ctx, 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("detail error = %v, want ErrNotFound", err)
	}
}
