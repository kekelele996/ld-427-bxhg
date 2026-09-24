package constants

// BudgetStatus 预算表状态。
type BudgetStatus string

const (
	BudgetStatusDraft    BudgetStatus = "Draft"
	BudgetStatusActive   BudgetStatus = "Active"
	BudgetStatusLocked   BudgetStatus = "Locked"
	BudgetStatusArchived BudgetStatus = "Archived"
)

// ExpenseStatus 支出审批状态。
type ExpenseStatus string

const (
	ExpenseStatusDraft     ExpenseStatus = "Draft"
	ExpenseStatusSubmitted ExpenseStatus = "Submitted"
	ExpenseStatusApproved  ExpenseStatus = "Approved"
	ExpenseStatusRejected  ExpenseStatus = "Rejected"
	ExpenseStatusPaid      ExpenseStatus = "Paid"
)

// PaymentMethod 支出方式。
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "Cash"
	PaymentMethodBankTransfer PaymentMethod = "BankTransfer"
	PaymentMethodCredit       PaymentMethod = "Credit"
	PaymentMethodCompany      PaymentMethod = "Company"
)

// SupplierStatus 供应商合作状态。
type SupplierStatus string

const (
	SupplierStatusActive      SupplierStatus = "Active"
	SupplierStatusSuspended   SupplierStatus = "Suspended"
	SupplierStatusBlacklisted SupplierStatus = "Blacklisted"
)

// SupplierCategory 供应商类别。
type SupplierCategory string

const (
	SupplierCategoryMaterial  SupplierCategory = "Material"
	SupplierCategoryFurniture SupplierCategory = "Furniture"
	SupplierCategoryAppliance SupplierCategory = "Appliance"
	SupplierCategoryLabor     SupplierCategory = "Labor"
	SupplierCategoryDesign    SupplierCategory = "Design"
	SupplierCategoryOther     SupplierCategory = "Other"
)

// BudgetCategory 预算类别。
type BudgetCategory string

const (
	BudgetCategoryDesign      BudgetCategory = "Design"
	BudgetCategoryMaterial    BudgetCategory = "Material"
	BudgetCategoryLabor       BudgetCategory = "Labor"
	BudgetCategoryFurniture   BudgetCategory = "Furniture"
	BudgetCategoryAppliance   BudgetCategory = "Appliance"
	BudgetCategoryContingency BudgetCategory = "Contingency"
	BudgetCategoryOther       BudgetCategory = "Other"
)

// ReconciliationStatus 对账状态。
type ReconciliationStatus string

const (
	ReconciliationStatusPending   ReconciliationStatus = "Pending"
	ReconciliationStatusConfirmed ReconciliationStatus = "Confirmed"
	ReconciliationStatusDisputed  ReconciliationStatus = "Disputed"
	ReconciliationStatusResolved  ReconciliationStatus = "Resolved"
)

// RoleName 角色名称。
type RoleName string

const (
	RoleAdmin          RoleName = "Admin"
	RoleFinanceManager RoleName = "FinanceManager"
	RoleProjectManager RoleName = "ProjectManager"
	RoleAccountant     RoleName = "Accountant"
	RoleOwner          RoleName = "Owner"
)

var (
	validBudgetStatuses = map[BudgetStatus]struct{}{
		BudgetStatusDraft: {}, BudgetStatusActive: {}, BudgetStatusLocked: {}, BudgetStatusArchived: {},
	}
	validExpenseStatuses = map[ExpenseStatus]struct{}{
		ExpenseStatusDraft: {}, ExpenseStatusSubmitted: {}, ExpenseStatusApproved: {}, ExpenseStatusRejected: {}, ExpenseStatusPaid: {},
	}
	validPaymentMethods = map[PaymentMethod]struct{}{
		PaymentMethodCash: {}, PaymentMethodBankTransfer: {}, PaymentMethodCredit: {}, PaymentMethodCompany: {},
	}
	validSupplierStatuses = map[SupplierStatus]struct{}{
		SupplierStatusActive: {}, SupplierStatusSuspended: {}, SupplierStatusBlacklisted: {},
	}
	validSupplierCategories = map[SupplierCategory]struct{}{
		SupplierCategoryMaterial: {}, SupplierCategoryFurniture: {}, SupplierCategoryAppliance: {}, SupplierCategoryLabor: {}, SupplierCategoryDesign: {}, SupplierCategoryOther: {},
	}
	validBudgetCategories = map[BudgetCategory]struct{}{
		BudgetCategoryDesign: {}, BudgetCategoryMaterial: {}, BudgetCategoryLabor: {}, BudgetCategoryFurniture: {}, BudgetCategoryAppliance: {}, BudgetCategoryContingency: {}, BudgetCategoryOther: {},
	}
	validReconciliationStatuses = map[ReconciliationStatus]struct{}{
		ReconciliationStatusPending: {}, ReconciliationStatusConfirmed: {}, ReconciliationStatusDisputed: {}, ReconciliationStatusResolved: {},
	}
)

func ValidBudgetStatus(s BudgetStatus) bool         { _, ok := validBudgetStatuses[s]; return ok }
func ValidExpenseStatus(s ExpenseStatus) bool       { _, ok := validExpenseStatuses[s]; return ok }
func ValidPaymentMethod(s PaymentMethod) bool       { _, ok := validPaymentMethods[s]; return ok }
func ValidSupplierStatus(s SupplierStatus) bool     { _, ok := validSupplierStatuses[s]; return ok }
func ValidSupplierCategory(s SupplierCategory) bool { _, ok := validSupplierCategories[s]; return ok }
func ValidBudgetCategory(s BudgetCategory) bool     { _, ok := validBudgetCategories[s]; return ok }
func ValidReconciliationStatus(s ReconciliationStatus) bool {
	_, ok := validReconciliationStatuses[s]
	return ok
}
