package repository

import (
	"context"

	"gorm.io/gorm"
)

type txContextKey struct{}

// TxManager 在单个数据库事务中执行多个仓储操作，保证跨表数据一致性。
type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type gormTxManager struct {
	db *gorm.DB
}

// NewTxManager 构造基于 GORM 的事务管理器。
func NewTxManager(db *gorm.DB) TxManager {
	return &gormTxManager{db: db}
}

func (m *gormTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txContextKey{}, tx))
	})
}

// connFor 返回当前上下文关联的事务句柄；未处于事务中时回退到传入的数据库句柄。
func connFor(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
