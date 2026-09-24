package constants

// 统一响应码。
const (
	CodeOK           = 0
	CodeBadRequest   = 40000
	CodeUnauthorized = 40100
	CodeForbidden    = 40300
	CodeNotFound     = 40400
	CodeConflict     = 40900
	CodeRateLimited  = 42900
	CodeInternal     = 50000
)

const MessageOK = "ok"
