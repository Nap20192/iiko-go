// Package customers covers loyalty: guests, cards, wallets, categories, coupons, programs and messaging.
package customers

import (
	"context"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// Service is the customers half of iikoCloud.
type Service struct{ c *rest.Client }

// New wires a service onto a transport; the facade is the usual caller.
func New(c *rest.Client) *Service { return &Service{c: c} }

// AddCustomerCard add card.
func (s *Service) AddCustomerCard(ctx context.Context, req gen.AddMagnetCardRequest) (*gen.AddMagnetCardResponse, error) {
	return rest.Call[gen.AddMagnetCardResponse](ctx, s.c, addCustomerCard, req)
}

// AddCustomerCategory add category for customer.
func (s *Service) AddCustomerCategory(ctx context.Context, req gen.ChangeCategoryForCustomerRequest) (*gen.ChangeCategoryForCustomerResponse, error) {
	return rest.Call[gen.ChangeCategoryForCustomerResponse](ctx, s.c, addCustomerCategory, req)
}

// AddCustomerProgram add customer to program.
func (s *Service) AddCustomerProgram(ctx context.Context, req gen.AddCustomerToProgramRequest) (*gen.AddCustomerToProgramResponse, error) {
	return rest.Call[gen.AddCustomerToProgramResponse](ctx, s.c, addCustomerProgram, req)
}

// CalculateLoyalty calculate checkin.
func (s *Service) CalculateLoyalty(ctx context.Context, req gen.CalculateCheckinRequest) (*gen.CalculateCheckinResponse, error) {
	return rest.Call[gen.CalculateCheckinResponse](ctx, s.c, calculateLoyalty, req)
}

// CancelHoldCustomerWallet cancel hold money.
func (s *Service) CancelHoldCustomerWallet(ctx context.Context, req gen.CancelHoldMoneyRequest) (*gen.CancelHoldMoneyResponse, error) {
	return rest.Call[gen.CancelHoldMoneyResponse](ctx, s.c, cancelHoldCustomerWallet, req)
}

// ChargeoffCustomerWallet withdraw balance.
func (s *Service) ChargeoffCustomerWallet(ctx context.Context, req gen.ChangeUserBalanceRequest) (*gen.WithdrawUserBalanceResponse, error) {
	return rest.Call[gen.WithdrawUserBalanceResponse](ctx, s.c, chargeoffCustomerWallet, req)
}

// CheckSMSSendingPossibility check sms sending possibility.
func (s *Service) CheckSMSSendingPossibility(ctx context.Context, req gen.SMSSendingPossibilityRequest) (*gen.SMSSendingPossibilityResponse, error) {
	return rest.Call[gen.SMSSendingPossibilityResponse](ctx, s.c, checkSMSSendingPossibility, req)
}

// CheckSMSStatus check SMS status.
func (s *Service) CheckSMSStatus(ctx context.Context, req gen.CheckSMSStatusRequest) (*gen.CheckSMSStatusResponse, error) {
	return rest.Call[gen.CheckSMSStatusResponse](ctx, s.c, checkSMSStatus, req)
}

// CreateOrUpdateCustomer create or update customer.
func (s *Service) CreateOrUpdateCustomer(ctx context.Context, req gen.CreateOrUpdateCustomerRequest) (*gen.CreateOrUpdateCustomerResponse, error) {
	return rest.Call[gen.CreateOrUpdateCustomerResponse](ctx, s.c, createOrUpdateCustomer, req)
}

// DeleteCustomers logical deletion of customers.
func (s *Service) DeleteCustomers(ctx context.Context, req gen.DeleteCustomersRequest) (*gen.DeleteCustomersResponse, error) {
	return rest.Call[gen.DeleteCustomersResponse](ctx, s.c, deleteCustomers, req)
}

// GetCounters get counters.
func (s *Service) GetCounters(ctx context.Context, req gen.GetCountersRequest) (*gen.GetCountersResponse, error) {
	return rest.Call[gen.GetCountersResponse](ctx, s.c, getCounters, req)
}

// GetCouponsBySeries get non-activated coupons.
func (s *Service) GetCouponsBySeries(ctx context.Context, req gen.NotActivatedCouponRequest) (*gen.NotActivatedCouponResponse, error) {
	return rest.Call[gen.NotActivatedCouponResponse](ctx, s.c, getCouponsBySeries, req)
}

// GetCouponsInfo get coupon info.
func (s *Service) GetCouponsInfo(ctx context.Context, req gen.CouponInfoRequest) (*gen.CouponInfoResponse, error) {
	return rest.Call[gen.CouponInfoResponse](ctx, s.c, getCouponsInfo, req)
}

// GetCouponsSeries get coupon series with non-activated coupons.
func (s *Service) GetCouponsSeries(ctx context.Context, req gen.SeriesWithNotActivatedCouponsRequest) (*gen.SeriesWithNotActivatedCouponsResponse, error) {
	return rest.Call[gen.SeriesWithNotActivatedCouponsResponse](ctx, s.c, getCouponsSeries, req)
}

// GetCustomerCategory get customer categories.
func (s *Service) GetCustomerCategory(ctx context.Context, req gen.GetCategoriesRequest) (*gen.GetCategoriesResponse, error) {
	return rest.Call[gen.GetCategoriesResponse](ctx, s.c, getCustomerCategory, req)
}

// GetCustomerInfo get customer info.
func (s *Service) GetCustomerInfo(ctx context.Context, req gen.GetCustomerInfoRequest) (*gen.GetCustomerInfoResponse, error) {
	return rest.Call[gen.GetCustomerInfoResponse](ctx, s.c, getCustomerInfo, req)
}

// GetCustomerTransactionsByDate get transaction report by period.
func (s *Service) GetCustomerTransactionsByDate(ctx context.Context, req gen.GetTransactionsReportByPeriodRequest) (*gen.GetTransactionsReportByPeriodResponse, error) {
	return rest.Call[gen.GetTransactionsReportByPeriodResponse](ctx, s.c, getCustomerTransactionsByDate, req)
}

// GetCustomerTransactionsByRevision get transaction report by revision.
func (s *Service) GetCustomerTransactionsByRevision(ctx context.Context, req gen.GetTransactionsReportByRevisionRequest) (*gen.GetTransactionsReportByRevisionResponse, error) {
	return rest.Call[gen.GetTransactionsReportByRevisionResponse](ctx, s.c, getCustomerTransactionsByRevision, req)
}

// GetManualCondition get manual conditions.
func (s *Service) GetManualCondition(ctx context.Context, req gen.GetByOrganizationIDRequest) (*gen.GetManualConditionsResponse, error) {
	return rest.Call[gen.GetManualConditionsResponse](ctx, s.c, getManualCondition, req)
}

// GetProgram get programs.
func (s *Service) GetProgram(ctx context.Context, req gen.GetProgramsRequest) (*gen.GetProgramsResponse, error) {
	return rest.Call[gen.GetProgramsResponse](ctx, s.c, getProgram, req)
}

// HoldCustomerWallet hold money.
func (s *Service) HoldCustomerWallet(ctx context.Context, req gen.HoldMoneyRequest) (*gen.HoldMoneyResponse, error) {
	return rest.Call[gen.HoldMoneyResponse](ctx, s.c, holdCustomerWallet, req)
}

// RemoveCustomerCard delete card.
func (s *Service) RemoveCustomerCard(ctx context.Context, req gen.DeleteMagnetCardRequest) (*gen.DeleteMagnetCardResponse, error) {
	return rest.Call[gen.DeleteMagnetCardResponse](ctx, s.c, removeCustomerCard, req)
}

// RemoveCustomerCategory remove category for customer.
func (s *Service) RemoveCustomerCategory(ctx context.Context, req gen.ChangeCategoryForCustomerRequest) (*gen.ChangeCategoryForCustomerResponse, error) {
	return rest.Call[gen.ChangeCategoryForCustomerResponse](ctx, s.c, removeCustomerCategory, req)
}

// RestoreCustomers logical recovery of customers.
func (s *Service) RestoreCustomers(ctx context.Context, req gen.RestoreCustomersRequest) (*gen.RestoreCustomersResponse, error) {
	return rest.Call[gen.RestoreCustomersResponse](ctx, s.c, restoreCustomers, req)
}

// SendEmailMessage send email.
func (s *Service) SendEmailMessage(ctx context.Context, req gen.SendEmailRequest) (*gen.SendEmailResponse, error) {
	return rest.Call[gen.SendEmailResponse](ctx, s.c, sendEmailMessage, req)
}

// SendSMSMessage send sms.
func (s *Service) SendSMSMessage(ctx context.Context, req gen.SendSMSRequest) (*gen.SendSMSResponse, error) {
	return rest.Call[gen.SendSMSResponse](ctx, s.c, sendSMSMessage, req)
}

// TopupCustomerWallet refill balance.
func (s *Service) TopupCustomerWallet(ctx context.Context, req gen.ChangeUserBalanceRequest) (*gen.RefillUserBalanceResponse, error) {
	return rest.Call[gen.RefillUserBalanceResponse](ctx, s.c, topupCustomerWallet, req)
}
