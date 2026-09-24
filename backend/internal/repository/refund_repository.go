package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/renovation/renovation-budget-api/internal/model"
)

// RefundRepository 支出退款数据访问接口。
type RefundRepository interface {
	Create(ctx context.Context, refund *model.ExpenseRefund) error
	FindByID(ctx context.Context, id uint) (*model.ExpenseRefund, error)
	ListByExpenseID(ctx context.Context, expenseID uint) ([]model.ExpenseRefund, error)
	ExistsByVoucherNo(ctx context.Context, voucherNo string) (bool, error)
	SumAmountByExpenseID(ctx context.Context, expenseID uint) (float64, error)
}

type refundRepository struct {
	db *gorm.DB
}

// NewRefundRepository 构造支出退款仓储。
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, refund *model.ExpenseRefund) error {
	err := r.db.WithContext(ctx).Create(refund).Error
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrDuplicateRefundVoucher
		}
		return fmt.Errorf("create expense refund: %w", err)
	}
	return nil
}

func (r *refundRepository) FindByID(ctx context.Context, id uint) (*model.ExpenseRefund, error) {
	var refund model.ExpenseRefund
	if err := r.db.WithContext(ctx).First(&refund, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find expense refund %d: %w", id, err)
	}
	return &refund, nil
}

func (r *refundRepository) ListByExpenseID(ctx context.Context, expenseID uint) ([]model.ExpenseRefund, error) {
	var refunds []model.ExpenseRefund
	if err := r.db.WithContext(ctx).
		Where("expense_id = ?", expenseID).
		Order("received_date DESC, id DESC").
		Find(&refunds).Error; err != nil {
		return nil, fmt.Errorf("list refunds for expense %d: %w", expenseID, err)
	}
	return refunds, nil
}

func (r *refundRepository) ExistsByVoucherNo(ctx context.Context, voucherNo string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ExpenseRefund{}).
		Where("voucher_no = ?", voucherNo).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check refund voucher number %q: %w", voucherNo, err)
	}
	return count > 0, nil
}

func (r *refundRepository) SumAmountByExpenseID(ctx context.Context, expenseID uint) (float64, error) {
	var total *float64
	if err := r.db.WithContext(ctx).Model(&model.ExpenseRefund{}).
		Where("expense_id = ?", expenseID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("sum refund amount for expense %d: %w", expenseID, err)
	}
	if total == nil {
		return 0, nil
	}
	return *total, nil
}

// isDuplicateKeyError 识别各驱动（PostgreSQL / SQLite）返回的唯一约束冲突。
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
