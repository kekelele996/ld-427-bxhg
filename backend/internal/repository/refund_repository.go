package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/model"
)

// ErrDuplicateVoucherNo 退款凭据号已存在，凭据号重复时只有第一次登记入账。
var ErrDuplicateVoucherNo = errors.New("repository: duplicate refund voucher no")

// RefundRepository 退款登记数据访问接口。
type RefundRepository interface {
	Create(ctx context.Context, record *model.RefundRecord) error
	FindByID(ctx context.Context, id uint) (*model.RefundRecord, error)
	ListByExpenseID(ctx context.Context, expenseID uint) ([]model.RefundRecord, error)
}

type refundRepository struct {
	db *gorm.DB
}

// NewRefundRepository 构造退款登记仓储。
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, record *model.RefundRecord) error {
	if err := connFor(ctx, r.db).Create(record).Error; err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicateVoucherNo
		}
		return fmt.Errorf("create refund record: %w", err)
	}
	return nil
}

func (r *refundRepository) FindByID(ctx context.Context, id uint) (*model.RefundRecord, error) {
	var record model.RefundRecord
	if err := connFor(ctx, r.db).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find refund record %d: %w", id, err)
	}
	return &record, nil
}

func (r *refundRepository) ListByExpenseID(ctx context.Context, expenseID uint) ([]model.RefundRecord, error) {
	var records []model.RefundRecord
	if err := connFor(ctx, r.db).Where("expense_id = ?", expenseID).Order("id ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list refund records for expense %d: %w", expenseID, err)
	}
	return records, nil
}

// isDuplicateKeyError 识别 PostgreSQL 与 SQLite（测试环境）的唯一约束冲突。
func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") || strings.Contains(msg, "duplicate key")
}
