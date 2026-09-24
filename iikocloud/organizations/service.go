// Package organizations covers the network-wide reference books: organizations, terminal groups, addresses and the dictionaries every other call needs an id from.
package organizations

import (
	"context"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// Service is the organizations half of iikoCloud.
type Service struct{ c *rest.Client }

// New wires a service onto a transport; the facade is the usual caller.
func New(c *rest.Client) *Service { return &Service{c: c} }

// AwakeTerminalGroups awake terminal groups from sleep mode.
func (s *Service) AwakeTerminalGroups(ctx context.Context, req gen.AwakeTerminalGroupsRequest) (*gen.AwakeTerminalGroupsResponse, error) {
	return rest.Call[gen.AwakeTerminalGroupsResponse](ctx, s.c, awakeTerminalGroups, req)
}

// GetCancelCauses delivery cancel causes.
func (s *Service) GetCancelCauses(ctx context.Context, req gen.CancelCausesRequest) (*gen.CancelCausesResponse, error) {
	return rest.Call[gen.CancelCausesResponse](ctx, s.c, getCancelCauses, req)
}

// GetCities cities.
func (s *Service) GetCities(ctx context.Context, req gen.CitiesRequest) (*gen.CitiesResponse, error) {
	return rest.Call[gen.CitiesResponse](ctx, s.c, getCities, req)
}

// GetCommandsStatus get status of command.
func (s *Service) GetCommandsStatus(ctx context.Context, req gen.GetCommandStatusRequest) (*gen.GetCommandStatusResponse, error) {
	return rest.Call[gen.GetCommandStatusResponse](ctx, s.c, getCommandsStatus, req)
}

// GetDeliveriesOrderTypes order types.
func (s *Service) GetDeliveriesOrderTypes(ctx context.Context, req gen.OrderTypesRequest) (*gen.OrderTypesResponse, error) {
	return rest.Call[gen.OrderTypesResponse](ctx, s.c, getDeliveriesOrderTypes, req)
}

// GetDeliveryRestrictions retrieve list of delivery restrictions.
func (s *Service) GetDeliveryRestrictions(ctx context.Context, req gen.GetDeliveryRestrictionsRequest) (*gen.GetDeliveryRestrictionsResponse, error) {
	return rest.Call[gen.GetDeliveryRestrictionsResponse](ctx, s.c, getDeliveryRestrictions, req)
}

// GetDeliveryRestrictionsAllowed get suitable terminal groups for delivery restrictions.
func (s *Service) GetDeliveryRestrictionsAllowed(ctx context.Context, req gen.GetAllowedRestrictionsRequest) (*gen.GetAllowedRestrictionsResponse, error) {
	return rest.Call[gen.GetAllowedRestrictionsResponse](ctx, s.c, getDeliveryRestrictionsAllowed, req)
}

// GetDiscounts discounts / surcharges.
func (s *Service) GetDiscounts(ctx context.Context, req gen.DiscountsRequest) (*gen.DiscountsResponse, error) {
	return rest.Call[gen.DiscountsResponse](ctx, s.c, getDiscounts, req)
}

// GetMarketingSources marketing sources.
func (s *Service) GetMarketingSources(ctx context.Context, req gen.MarketingSourcesRequest) (*gen.MarketingSourcesResponse, error) {
	return rest.Call[gen.MarketingSourcesResponse](ctx, s.c, getMarketingSources, req)
}

// GetOrganizations returns organizations available to api-login user.
func (s *Service) GetOrganizations(ctx context.Context, req gen.GetOrganizationsRequest) (*gen.GetOrganizationsResponse, error) {
	return rest.Call[gen.GetOrganizationsResponse](ctx, s.c, getOrganizations, req)
}

// GetOrganizationsSettings returns available to api-login user organizations specified settings.
func (s *Service) GetOrganizationsSettings(ctx context.Context, req gen.OrganizationsSettingsRequest) (*gen.OrganizationsSettingsResponse, error) {
	return rest.Call[gen.OrganizationsSettingsResponse](ctx, s.c, getOrganizationsSettings, req)
}

// GetPaymentTypes payment types.
func (s *Service) GetPaymentTypes(ctx context.Context, req gen.PaymentTypesRequest) (*gen.PaymentTypesResponse, error) {
	return rest.Call[gen.PaymentTypesResponse](ctx, s.c, getPaymentTypes, req)
}

// GetRegions regions.
func (s *Service) GetRegions(ctx context.Context, req gen.RegionsRequest) (*gen.RegionsResponse, error) {
	return rest.Call[gen.RegionsResponse](ctx, s.c, getRegions, req)
}

// GetRemovalTypes removal types (reasons for deletion).
func (s *Service) GetRemovalTypes(ctx context.Context, req gen.RemovalTypesRequest) (*gen.RemovalTypesResponse, error) {
	return rest.Call[gen.RemovalTypesResponse](ctx, s.c, getRemovalTypes, req)
}

// GetStreetsByCity streets by city.
func (s *Service) GetStreetsByCity(ctx context.Context, req gen.StreetsByCityRequest) (*gen.StreetsResponse, error) {
	return rest.Call[gen.StreetsResponse](ctx, s.c, getStreetsByCity, req)
}

// GetStreetsByID streets by id or by classifierId.
func (s *Service) GetStreetsByID(ctx context.Context, req gen.StreetsByIDRequest) (*gen.StreetsByIDResponse, error) {
	return rest.Call[gen.StreetsByIDResponse](ctx, s.c, getStreetsByID, req)
}

// GetTerminalGroups method that returns information on groups of delivery terminals.
func (s *Service) GetTerminalGroups(ctx context.Context, req gen.TerminalGroupsRequest) (*gen.TerminalGroupsResponse, error) {
	return rest.Call[gen.TerminalGroupsResponse](ctx, s.c, getTerminalGroups, req)
}

// GetTerminalGroupsIsAlive returns information on availability of group of terminals.
func (s *Service) GetTerminalGroupsIsAlive(ctx context.Context, req gen.TerminalGroupsIsAliveRequest) (*gen.TerminalGroupsIsAliveResponse, error) {
	return rest.Call[gen.TerminalGroupsIsAliveResponse](ctx, s.c, getTerminalGroupsIsAlive, req)
}

// GetTipsTypes get tips types for api-login`s rms group.
func (s *Service) GetTipsTypes(ctx context.Context) (*gen.TipsTypesResponse, error) {
	return rest.Call[gen.TipsTypesResponse](ctx, s.c, getTipsTypes, nil)
}

// ListLicenses get license list information for the API login.
func (s *Service) ListLicenses(ctx context.Context, req gen.GetLicenseListRequest) (*gen.GetLicenseListResponse, error) {
	return rest.Call[gen.GetLicenseListResponse](ctx, s.c, listLicenses, req)
}

// SendNotifications send notification to external systems.
func (s *Service) SendNotifications(ctx context.Context, req gen.SendNotificationRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, sendNotifications, req)
}
