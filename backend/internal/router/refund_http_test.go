package router

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/config"
	"github.com/renovation/renovation-budget-api/internal/handler"
	"github.com/renovation/renovation-budget-api/internal/model"
	"github.com/renovation/renovation-budget-api/internal/repository"
	"github.com/renovation/renovation-budget-api/internal/service"
)

type apiResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func buildTestEngine(t *testing.T) (*ginEngine, string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_busy_timeout=5000"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Role{}, &model.User{}, &model.BudgetSheet{}, &model.BudgetItem{},
		&model.ExpenseRecord{}, &model.RefundRecord{}, &model.Supplier{},
		&model.Reconciliation{}, &model.AuditLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := repository.SeedInitialData(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	budgetRepo := repository.NewBudgetRepository(db)
	itemRepo := repository.NewItemRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	refundRepo := repository.NewRefundRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	reconciliationRepo := repository.NewReconciliationRepository(db)
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	cfg := &config.Config{JWTSecret: "test-secret", JWTExpireHours: 24, RateLimitMax: 100, RateLimitWindowSeconds: 60}

	auditSvc := service.NewAuditService(auditRepo, logger)
	authSvc := service.NewAuthService(userRepo, cfg, logger)
	budgetSvc := service.NewBudgetService(budgetRepo, itemRepo, auditSvc, nil, logger)
	itemSvc := service.NewItemService(itemRepo, budgetRepo, auditSvc, nil, logger)
	expenseSvc := service.NewExpenseService(expenseRepo, itemRepo, budgetRepo, auditSvc, nil, logger)
	refundSvc := service.NewRefundService(repository.NewTxManager(db), refundRepo, expenseRepo, itemRepo, budgetRepo, auditSvc, nil, logger)
	supplierSvc := service.NewSupplierService(supplierRepo, auditSvc, logger)
	reconciliationSvc := service.NewReconciliationService(reconciliationRepo, auditSvc, logger)

	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}) // 不可达：限流中间件会放行
	engine := New(Dependencies{
		Config: cfg, Redis: rdb, Logger: logger, AuditRepo: auditRepo,
		AuthHandler:           handler.NewAuthHandler(authSvc, logger),
		AuditHandler:          handler.NewAuditHandler(auditSvc, logger),
		BudgetHandler:         handler.NewBudgetHandler(budgetSvc, logger),
		ItemHandler:           handler.NewItemHandler(itemSvc, logger),
		ExpenseHandler:        handler.NewExpenseHandler(expenseSvc, refundSvc, logger),
		RefundHandler:         handler.NewRefundHandler(refundSvc, logger),
		SupplierHandler:       handler.NewSupplierHandler(supplierSvc, logger),
		ReconciliationHandler: handler.NewReconciliationHandler(reconciliationSvc, logger),
	})
	return &ginEngine{engine}, ""
}

type ginEngine struct{ e http.Handler }

func (g *ginEngine) do(t *testing.T, method, path, token string, body any) (apiResp, *httptest.ResponseRecorder) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		raw, _ := json.Marshal(body)
		buf.Write(raw)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	g.e.ServeHTTP(rec, req)
	var resp apiResp
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response %s: %v\nbody: %s", path, err, rec.Body.String())
		}
	}
	return resp, rec
}

func loginToken(t *testing.T, g *ginEngine, username, password string) string {
	t.Helper()
	resp, rec := g.do(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"username": username, "password": password})
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s status=%d", username, rec.Code)
	}
	var data struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	return data.Token
}

func dataUint(t *testing.T, raw json.RawMessage, key string) uint {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	var v uint
	if err := json.Unmarshal(m[key], &v); err != nil {
		t.Fatalf("unmarshal %s: %v", key, err)
	}
	return v
}

func TestRefundHTTPEndToEnd(t *testing.T) {
	g, _ := buildTestEngine(t)
	finance := loginToken(t, g, "finance", "Finance123!")
	owner := loginToken(t, g, "owner", "Owner123!")

	// 预算 → 分项
	resp, rec := g.do(t, http.MethodPost, "/api/v1/budgets", finance, map[string]any{
		"project_id": "p-http", "name": "HTTP 装修预算", "total_amount": 8000, "status": "Active",
	})
	if rec.Code != http.StatusOK || resp.Code != 0 {
		t.Fatalf("create budget status=%d code=%d msg=%s", rec.Code, resp.Code, resp.Message)
	}
	budgetID := dataUint(t, resp.Data, "id")

	resp, rec = g.do(t, http.MethodPost, "/api/v1/budgets/"+itoa(budgetID)+"/items", finance, map[string]any{
		"category": "Material", "sub_category": "瓷砖", "budget_amount": 3000,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("create item status=%d body=%s", rec.Code, rec.Body.String())
	}
	itemID := dataUint(t, resp.Data, "id")

	// 支出 → 提交 → 审批 → 付款
	resp, rec = g.do(t, http.MethodPost, "/api/v1/expenses", finance, map[string]any{
		"budget_item_id": itemID, "amount": 1000, "expense_date": "2026-09-01",
		"payment_method": "BankTransfer", "description": "瓷砖采购",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("create expense status=%d body=%s", rec.Code, rec.Body.String())
	}
	expenseID := dataUint(t, resp.Data, "id")

	if _, rec := g.do(t, http.MethodPost, "/api/v1/expenses/"+itoa(expenseID)+"/submit", finance, nil); rec.Code != http.StatusOK {
		t.Fatalf("submit status=%d body=%s", rec.Code, rec.Body.String())
	}
	if _, rec := g.do(t, http.MethodPost, "/api/v1/expenses/"+itoa(expenseID)+"/approve", finance,
		map[string]string{"approval_comment": "同意"}); rec.Code != http.StatusOK {
		t.Fatalf("approve status=%d body=%s", rec.Code, rec.Body.String())
	}
	if _, rec := g.do(t, http.MethodPost, "/api/v1/expenses/"+itoa(expenseID)+"/pay", finance,
		map[string]string{"payment_date": "2026-09-05"}); rec.Code != http.StatusOK {
		t.Fatalf("pay status=%d body=%s", rec.Code, rec.Body.String())
	}

	// 业主只读：不能登记退款。
	if _, rec := g.do(t, http.MethodPost, "/api/v1/expenses/"+itoa(expenseID)+"/refunds", owner, map[string]any{
		"amount": 100, "received_date": "2026-09-10", "voucher_no": "R-OWNER",
	}); rec.Code != http.StatusForbidden {
		t.Fatalf("owner refund status=%d, want 403", rec.Code)
	}

	// 参数校验：金额必须大于 0。
	if resp, rec := g.do(t, http.MethodPost, "/api/v1/expenses/"+itoa(expenseID)+"/refunds", finance, map[string]any{
		"amount": 0, "received_date": "2026-09-10", "voucher_no": "R-BAD",
	}); rec.Code != http.StatusBadRequest || resp.Code != 40000 {
		t.Fatalf("invalid amount status=%d code=%d", rec.Code, resp.Code)
	}

	// 第一笔退款 400。
	if _, rec := g.do(t, http.MethodPost, "/api/v1/expenses/"+itoa(expenseID)+"/refunds", finance, map[string]any{
		"amount": 400, "received_date": "2026-09-10", "reason": "瓷砖退料", "voucher_no": "R-0001",
	}); rec.Code != http.StatusOK {
		t.Fatalf("refund #1 status=%d body=%s", rec.Code, rec.Body.String())
	}

	// 超额退款 700：409，且分项/预算不变。
	if resp, rec := g.do(t, http.MethodPost, "/api/v1/expenses/"+itoa(expenseID)+"/refunds", finance, map[string]any{
		"amount": 700, "received_date": "2026-09-11", "reason": "超额", "voucher_no": "R-0002",
	}); rec.Code != http.StatusConflict || resp.Code != 40900 {
		t.Fatalf("exceed refund status=%d code=%d body=%s", rec.Code, resp.Code, rec.Body.String())
	}

	// 重复凭据号：409，只入账第一次。
	if resp, rec := g.do(t, http.MethodPost, "/api/v1/expenses/"+itoa(expenseID)+"/refunds", finance, map[string]any{
		"amount": 100, "received_date": "2026-09-12", "voucher_no": "R-0001",
	}); rec.Code != http.StatusConflict || resp.Code != 40900 {
		t.Fatalf("duplicate voucher status=%d code=%d", rec.Code, resp.Code)
	}

	// 退满剩余 600。
	if _, rec := g.do(t, http.MethodPost, "/api/v1/expenses/"+itoa(expenseID)+"/refunds", finance, map[string]any{
		"amount": 600, "received_date": "2026-09-13", "reason": "减项", "voucher_no": "R-0003",
	}); rec.Code != http.StatusOK {
		t.Fatalf("refund #3 status=%d body=%s", rec.Code, rec.Body.String())
	}

	// 支出详情：已退 1000、剩余可退 0、两条退款记录。
	resp, rec = g.do(t, http.MethodGet, "/api/v1/expenses/"+itoa(expenseID), finance, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get expense status=%d", rec.Code)
	}
	var detail struct {
		RefundedAmount float64 `json:"refunded_amount"`
		Refund         struct {
			RefundedAmount float64 `json:"refunded_amount"`
			RefundableLeft float64 `json:"refundable_left"`
			Refunds        []struct {
				VoucherNo string  `json:"voucher_no"`
				Amount    float64 `json:"amount"`
			} `json:"refunds"`
		} `json:"refund"`
	}
	if err := json.Unmarshal(resp.Data, &detail); err != nil {
		t.Fatalf("decode detail: %v", err)
	}
	if detail.RefundedAmount != 1000 || detail.Refund.RefundedAmount != 1000 || detail.Refund.RefundableLeft != 0 {
		t.Fatalf("detail amounts: %+v", detail)
	}
	if len(detail.Refund.Refunds) != 2 {
		t.Fatalf("refund records = %d, want 2", len(detail.Refund.Refunds))
	}

	// 退款记录列表接口。
	resp, rec = g.do(t, http.MethodGet, "/api/v1/expenses/"+itoa(expenseID)+"/refunds", finance, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list refunds status=%d", rec.Code)
	}
	var list []struct {
		VoucherNo string `json:"voucher_no"`
	}
	if err := json.Unmarshal(resp.Data, &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2", len(list))
	}

	// 预算表已用金额回冲为 0、可用余额恢复为 8000。
	resp, rec = g.do(t, http.MethodGet, "/api/v1/budgets/"+itoa(budgetID), finance, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get budget status=%d", rec.Code)
	}
	var budget struct {
		SpentAmount     float64 `json:"spent_amount"`
		AvailableAmount float64 `json:"available_amount"`
	}
	if err := json.Unmarshal(resp.Data, &budget); err != nil {
		t.Fatalf("decode budget: %v", err)
	}
	if budget.SpentAmount != 0 || budget.AvailableAmount != 8000 {
		t.Fatalf("budget spent=%v available=%v, want 0/8000", budget.SpentAmount, budget.AvailableAmount)
	}

	// 不存在的支出登记退款返回 404。
	if resp, rec := g.do(t, http.MethodPost, "/api/v1/expenses/9999/refunds", finance, map[string]any{
		"amount": 1, "received_date": "2026-09-13", "voucher_no": "R-X",
	}); rec.Code != http.StatusNotFound || resp.Code != 40400 {
		t.Fatalf("missing expense status=%d code=%d", rec.Code, resp.Code)
	}
}

func itoa(v uint) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
