// Package finance covers accounts, chart of accounts, transactions and service documents.
package finance

import (
	"context"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// Service is the finance half of iikoCloud.
type Service struct{ c *rest.Client }

// New wires a service onto a transport; the facade is the usual caller.
func New(c *rest.Client) *Service { return &Service{c: c} }

// CancelIncomingService cancel incoming service act draft.
func (s *Service) CancelIncomingService(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingServiceSaveResponse, error) {
	return rest.Call[gen.IncomingServiceSaveResponse](ctx, s.c, cancelIncomingService, req)
}

// CancelOutgoingService cancel outgoing service act draft.
func (s *Service) CancelOutgoingService(ctx context.Context, req gen.GetByIDRequest) (*gen.OutgoingServiceSaveResponse, error) {
	return rest.Call[gen.OutgoingServiceSaveResponse](ctx, s.c, cancelOutgoingService, req)
}

// CreateAccount create financial account.
func (s *Service) CreateAccount(ctx context.Context, req gen.AccountCreateRequest) error {
	return s.c.Post(ctx, createAccount, req, nil)
}

// CreateCashFlowCategory create cash flow category.
func (s *Service) CreateCashFlowCategory(ctx context.Context, req gen.CashFlowCategoryCreateRequest) error {
	return s.c.Post(ctx, createCashFlowCategory, req, nil)
}

// CreateIncomingService create incoming service act.
func (s *Service) CreateIncomingService(ctx context.Context, req gen.IncomingServiceCreateRequest) error {
	return s.c.Post(ctx, createIncomingService, req, nil)
}

// CreateOutgoingService create outgoing service act.
func (s *Service) CreateOutgoingService(ctx context.Context, req gen.OutgoingServiceCreateRequest) error {
	return s.c.Post(ctx, createOutgoingService, req, nil)
}

// DeleteAccount delete financial account.
func (s *Service) DeleteAccount(ctx context.Context, req gen.AccountDeleteRequest) (*gen.AccountDeleteResponse, error) {
	return rest.Call[gen.AccountDeleteResponse](ctx, s.c, deleteAccount, req)
}

// DeleteCashFlowCategory delete cash flow category.
func (s *Service) DeleteCashFlowCategory(ctx context.Context, req gen.CashFlowCategoryDeleteRequest) (*gen.CashFlowCategoryCascadeResponse, error) {
	return rest.Call[gen.CashFlowCategoryCascadeResponse](ctx, s.c, deleteCashFlowCategory, req)
}

// GetAccount get financial account.
func (s *Service) GetAccount(ctx context.Context, req gen.AccountGetRequest) (*gen.AccountResponse, error) {
	return rest.Call[gen.AccountResponse](ctx, s.c, getAccount, req)
}

// GetCashFlowCategory get cash flow category.
func (s *Service) GetCashFlowCategory(ctx context.Context, req gen.CashFlowCategoryGetRequest) (*gen.CashFlowCategoryResponse, error) {
	return rest.Call[gen.CashFlowCategoryResponse](ctx, s.c, getCashFlowCategory, req)
}

// GetIncomingService get incoming service act.
func (s *Service) GetIncomingService(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingServiceGetResponse, error) {
	return rest.Call[gen.IncomingServiceGetResponse](ctx, s.c, getIncomingService, req)
}

// GetOutgoingService get outgoing service act.
func (s *Service) GetOutgoingService(ctx context.Context, req gen.GetByIDRequest) (*gen.OutgoingServiceGetResponse, error) {
	return rest.Call[gen.OutgoingServiceGetResponse](ctx, s.c, getOutgoingService, req)
}

// ListAccount list of financial accounts.
func (s *Service) ListAccount(ctx context.Context, req gen.AccountListRequest) (*gen.AccountListResponse, error) {
	return rest.Call[gen.AccountListResponse](ctx, s.c, listAccount, req)
}

// ListAccountPosting account postings.
func (s *Service) ListAccountPosting(ctx context.Context, req gen.AccountPostingListRequest) (*gen.AccountPostingListResponse, error) {
	return rest.Call[gen.AccountPostingListResponse](ctx, s.c, listAccountPosting, req)
}

// ListAccountTransactions get account transactions.
func (s *Service) ListAccountTransactions(ctx context.Context, req gen.AccountTransactionsListRequest) (*gen.AccountTransactionsResponse, error) {
	return rest.Call[gen.AccountTransactionsResponse](ctx, s.c, listAccountTransactions, req)
}

// ListAccountType list of account types.
func (s *Service) ListAccountType(ctx context.Context, req gen.AccountTypeListRequest) ([]gen.AccountType, error) {
	return rest.CallList[gen.AccountType](ctx, s.c, listAccountType, req)
}

// ListBalanceSheet balance sheet.
func (s *Service) ListBalanceSheet(ctx context.Context, req gen.BalanceSheetListRequest) (*gen.BalanceSheetListResponse, error) {
	return rest.Call[gen.BalanceSheetListResponse](ctx, s.c, listBalanceSheet, req)
}

// ListCashFlowCategory list of cash flow categories.
func (s *Service) ListCashFlowCategory(ctx context.Context, req gen.CashFlowCategoryListRequest) (*gen.CashFlowCategoryListResponse, error) {
	return rest.Call[gen.CashFlowCategoryListResponse](ctx, s.c, listCashFlowCategory, req)
}

// ListChartOfAccounts chart of accounts.
func (s *Service) ListChartOfAccounts(ctx context.Context, req gen.ChartOfAccountsListRequest) (*gen.ChartOfAccountsListResponse, error) {
	return rest.Call[gen.ChartOfAccountsListResponse](ctx, s.c, listChartOfAccounts, req)
}

// ListDocumentTransactions get document transactions.
func (s *Service) ListDocumentTransactions(ctx context.Context, req gen.DocumentTransactionsListRequest) ([]gen.DocumentTransactionItem, error) {
	return rest.CallList[gen.DocumentTransactionItem](ctx, s.c, listDocumentTransactions, req)
}

// ListIncomingService export incoming service acts.
func (s *Service) ListIncomingService(ctx context.Context, req gen.ListRequest) ([]gen.IncomingServiceListItem, error) {
	return rest.CallList[gen.IncomingServiceListItem](ctx, s.c, listIncomingService, req)
}

// ListItemCategory get a list of fiscal categories.
func (s *Service) ListItemCategory(ctx context.Context, req gen.ItemCategoryListRequest) (*gen.ItemCategoryListResponse, error) {
	return rest.Call[gen.ItemCategoryListResponse](ctx, s.c, listItemCategory, req)
}

// ListOutgoingService export outgoing service acts.
func (s *Service) ListOutgoingService(ctx context.Context, req gen.ListRequest) ([]gen.OutgoingServiceListItem, error) {
	return rest.CallList[gen.OutgoingServiceListItem](ctx, s.c, listOutgoingService, req)
}

// ListTaxCategory get a list of tax categories.
func (s *Service) ListTaxCategory(ctx context.Context, req gen.TaxCategoryListRequest) (*gen.TaxCategoryListResponse, error) {
	return rest.Call[gen.TaxCategoryListResponse](ctx, s.c, listTaxCategory, req)
}

// PostIncomingService post incoming service act.
func (s *Service) PostIncomingService(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingServiceSaveResponse, error) {
	return rest.Call[gen.IncomingServiceSaveResponse](ctx, s.c, postIncomingService, req)
}

// PostOutgoingService post outgoing service act.
func (s *Service) PostOutgoingService(ctx context.Context, req gen.GetByIDRequest) (*gen.OutgoingServiceSaveResponse, error) {
	return rest.Call[gen.OutgoingServiceSaveResponse](ctx, s.c, postOutgoingService, req)
}

// RestoreAccount restore financial account.
func (s *Service) RestoreAccount(ctx context.Context, req gen.AccountRestoreRequest) (*gen.AccountRestoreResponse, error) {
	return rest.Call[gen.AccountRestoreResponse](ctx, s.c, restoreAccount, req)
}

// RestoreCashFlowCategory restore cash flow category.
func (s *Service) RestoreCashFlowCategory(ctx context.Context, req gen.CashFlowCategoryRestoreRequest) (*gen.CashFlowCategoryCascadeResponse, error) {
	return rest.Call[gen.CashFlowCategoryCascadeResponse](ctx, s.c, restoreCashFlowCategory, req)
}

// UnpostIncomingService unpost incoming service act.
func (s *Service) UnpostIncomingService(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingServiceSaveResponse, error) {
	return rest.Call[gen.IncomingServiceSaveResponse](ctx, s.c, unpostIncomingService, req)
}

// UnpostOutgoingService unpost outgoing service act.
func (s *Service) UnpostOutgoingService(ctx context.Context, req gen.GetByIDRequest) (*gen.OutgoingServiceSaveResponse, error) {
	return rest.Call[gen.OutgoingServiceSaveResponse](ctx, s.c, unpostOutgoingService, req)
}

// UpdateAccount update financial account.
func (s *Service) UpdateAccount(ctx context.Context, req gen.AccountUpdateRequest) (*gen.AccountResponse, error) {
	return rest.Call[gen.AccountResponse](ctx, s.c, updateAccount, req)
}

// UpdateCashFlowCategory update cash flow category.
func (s *Service) UpdateCashFlowCategory(ctx context.Context, req gen.CashFlowCategoryUpdateRequest) (*gen.CashFlowCategoryResponse, error) {
	return rest.Call[gen.CashFlowCategoryResponse](ctx, s.c, updateCashFlowCategory, req)
}

// UpdateIncomingService edit incoming service act.
func (s *Service) UpdateIncomingService(ctx context.Context, req gen.IncomingServiceUpdateRequest) (*gen.IncomingServiceSaveResponse, error) {
	return rest.Call[gen.IncomingServiceSaveResponse](ctx, s.c, updateIncomingService, req)
}

// UpdateOutgoingService edit outgoing service act.
func (s *Service) UpdateOutgoingService(ctx context.Context, req gen.OutgoingServiceUpdateRequest) (*gen.OutgoingServiceSaveResponse, error) {
	return rest.Call[gen.OutgoingServiceSaveResponse](ctx, s.c, updateOutgoingService, req)
}
