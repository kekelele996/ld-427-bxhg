package service

import "errors"

// 服务层哨兵错误。
var (
	ErrNotFound            = errors.New("service: not found")
	ErrInvalidLogin        = errors.New("service: invalid username or password")
	ErrInvalidState        = errors.New("service: invalid state transition")
	ErrInsufficientBalance = errors.New("service: insufficient available balance")
	ErrForbiddenTransition = errors.New("service: forbidden transition")
	ErrRefundExceedsPaid   = errors.New("service: accumulated refund exceeds paid amount")
	ErrDuplicateVoucherNo  = errors.New("service: refund voucher number already registered")
	ErrRefundNotAllowed    = errors.New("service: refund only allowed for paid expense")
)
