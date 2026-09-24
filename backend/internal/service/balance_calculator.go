package service

// CalculateAvailable 计算预算表可用余额。
func CalculateAvailable(totalAmount, spentAmount, frozenAmount float64) float64 {
	return totalAmount - spentAmount - frozenAmount
}

// CalculateVariance 计算预算项差异金额。
func CalculateVariance(spentAmount, budgetAmount float64) float64 {
	return spentAmount - budgetAmount
}

// CalculateUnpaid 计算对账单未付金额。
func CalculateUnpaid(payableAmount, paidAmount float64) float64 {
	return payableAmount - paidAmount
}
