// Package orders covers table orders, deliveries, delivery drafts and banquet reserves.
package orders

import (
	"context"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// Service is the orders half of iikoCloud.
type Service struct{ c *rest.Client }

// New wires a service onto a transport; the facade is the usual caller.
func New(c *rest.Client) *Service { return &Service{c: c} }

// AddCustomerOrder add customer to order.
func (s *Service) AddCustomerOrder(ctx context.Context, req gen.AddCustomerToTableOrderRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, addCustomerOrder, req)
}

// AddItemsDeliveries add order items.
func (s *Service) AddItemsDeliveries(ctx context.Context, req gen.AddOrderItemsRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, addItemsDeliveries, req)
}

// AddItemsOrder add order items.
func (s *Service) AddItemsOrder(ctx context.Context, req gen.AddItemsToTableOrderRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, addItemsOrder, req)
}

// AddItemsReserve add order items.
func (s *Service) AddItemsReserve(ctx context.Context, req gen.AddOrderItemsToBanquetRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, addItemsReserve, req)
}

// AddPaymentsDeliveries add order payments.
func (s *Service) AddPaymentsDeliveries(ctx context.Context, req gen.AddOrderPaymentsRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, addPaymentsDeliveries, req)
}

// AddPaymentsOrder add order payments.
func (s *Service) AddPaymentsOrder(ctx context.Context, req gen.AddOrderPaymentsRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, addPaymentsOrder, req)
}

// AddPaymentsReserve add order payments.
func (s *Service) AddPaymentsReserve(ctx context.Context, req gen.AddOrderPaymentsToBanquetRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, addPaymentsReserve, req)
}

// CancelConfirmationDeliveries cancel delivery confirmation.
func (s *Service) CancelConfirmationDeliveries(ctx context.Context, req gen.CancelDeliveryConfirmationRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, cancelConfirmationDeliveries, req)
}

// CancelDeliveries cancel delivery order.
func (s *Service) CancelDeliveries(ctx context.Context, req gen.CancelOrderRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, cancelDeliveries, req)
}

// CancelOrder cancel the table order.
func (s *Service) CancelOrder(ctx context.Context, req gen.CancelTableOrderRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, cancelOrder, req)
}

// CancelReserve cancel reservation due to some reason.
func (s *Service) CancelReserve(ctx context.Context, req gen.CancelReserveRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, cancelReserve, req)
}

// ChangeCommentDeliveries change delivery comment.
func (s *Service) ChangeCommentDeliveries(ctx context.Context, req gen.ChangeDeliveryCommentRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeCommentDeliveries, req)
}

// ChangeCompleteBeforeDeliveries change time when client wants the order to be delivered.
func (s *Service) ChangeCompleteBeforeDeliveries(ctx context.Context, req gen.ChangeCompleteBeforeRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeCompleteBeforeDeliveries, req)
}

// ChangeDeliveryPointDeliveries change order's delivery point information.
func (s *Service) ChangeDeliveryPointDeliveries(ctx context.Context, req gen.ChangeDeliveryPointRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeDeliveryPointDeliveries, req)
}

// ChangeDriverInfoDeliveries change driver info.
func (s *Service) ChangeDriverInfoDeliveries(ctx context.Context, req gen.ChangeDriverInfoRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeDriverInfoDeliveries, req)
}

// ChangeEstimatedStartTimeReserve change reserve/banquet estimated start time.
func (s *Service) ChangeEstimatedStartTimeReserve(ctx context.Context, req gen.ChangeReserveEstimatedStartTimeRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeEstimatedStartTimeReserve, req)
}

// ChangeExternalDataDeliveries change delivery external data.
func (s *Service) ChangeExternalDataDeliveries(ctx context.Context, req gen.ChangeExternalDataRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeExternalDataDeliveries, req)
}

// ChangeExternalDataOrder change table order external_data.
func (s *Service) ChangeExternalDataOrder(ctx context.Context, req gen.ChangeExternalDataRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeExternalDataOrder, req)
}

// ChangeItemsReserve change order items.
func (s *Service) ChangeItemsReserve(ctx context.Context, req gen.ChangeBanquetOrderItemsRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeItemsReserve, req)
}

// ChangeOperatorDeliveries assign/change the order operator.
func (s *Service) ChangeOperatorDeliveries(ctx context.Context, req gen.ChangeDeliveryOperatorRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeOperatorDeliveries, req)
}

// ChangePaymentsDeliveries change order's payments.
func (s *Service) ChangePaymentsDeliveries(ctx context.Context, req gen.ChangePaymentsRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changePaymentsDeliveries, req)
}

// ChangePaymentsOrder change table order's payments.
func (s *Service) ChangePaymentsOrder(ctx context.Context, req gen.ChangePaymentsRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changePaymentsOrder, req)
}

// ChangeServiceTypeDeliveries change order's delivery type.
func (s *Service) ChangeServiceTypeDeliveries(ctx context.Context, req gen.ChangeServiceTypeRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeServiceTypeDeliveries, req)
}

// ChangeTablesReserve change reserve/banquet tables.
func (s *Service) ChangeTablesReserve(ctx context.Context, req gen.ChangeReserveTablesRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, changeTablesReserve, req)
}

// CloseDeliveries close order.
func (s *Service) CloseDeliveries(ctx context.Context, req gen.CloseDeliveryOrderRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, closeDeliveries, req)
}

// CloseOrder close order.
func (s *Service) CloseOrder(ctx context.Context, req gen.CloseTableOrderRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, closeOrder, req)
}

// CommitDeliveriesDrafts admit order draft changes and send them to Front.
func (s *Service) CommitDeliveriesDrafts(ctx context.Context, req gen.CommitDraftRequest) (*gen.OrderResponse, error) {
	return rest.Call[gen.OrderResponse](ctx, s.c, commitDeliveriesDrafts, req)
}

// ConfirmDeliveries confirm delivery.
func (s *Service) ConfirmDeliveries(ctx context.Context, req gen.ConfirmDeliveryRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, confirmDeliveries, req)
}

// CreateDeliveries create delivery.
func (s *Service) CreateDeliveries(ctx context.Context, req gen.CreateOrderRequest) (*gen.OrderResponse, error) {
	return rest.Call[gen.OrderResponse](ctx, s.c, createDeliveries, req)
}

// CreateDeliveriesDrafts create delivery order draft.
func (s *Service) CreateDeliveriesDrafts(ctx context.Context, req gen.CreateDraftRequest) (*gen.CreateOrSaveDraftResponse, error) {
	return rest.Call[gen.CreateOrSaveDraftResponse](ctx, s.c, createDeliveriesDrafts, req)
}

// CreateOrder create order.
func (s *Service) CreateOrder(ctx context.Context, req gen.CreateTableOrderRequest) (*gen.TableOrderResponse, error) {
	return rest.Call[gen.TableOrderResponse](ctx, s.c, createOrder, req)
}

// CreateReserve create banquet/reserve.
func (s *Service) CreateReserve(ctx context.Context, req gen.CreateReserveRequest) (*gen.ReserveResponse, error) {
	return rest.Call[gen.ReserveResponse](ctx, s.c, createReserve, req)
}

// DeleteDeliveriesDrafts delete order draft.
func (s *Service) DeleteDeliveriesDrafts(ctx context.Context, req gen.DeleteDraftRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, deleteDeliveriesDrafts, req)
}

// GetDeliveriesByDeliveryDateAndPhone retrieve list of orders by telephone number, dates and revision.
func (s *Service) GetDeliveriesByDeliveryDateAndPhone(ctx context.Context, req gen.OrdersByDeliveryDateAndPhoneRequest) (*gen.OrdersWithRevisionResponse, error) {
	return rest.Call[gen.OrdersWithRevisionResponse](ctx, s.c, getDeliveriesByDeliveryDateAndPhone, req)
}

// GetDeliveriesByDeliveryDateAndSourceKeyAndFilter search orders by search text and additional filters (date, problem, statuses and other).
func (s *Service) GetDeliveriesByDeliveryDateAndSourceKeyAndFilter(ctx context.Context, req gen.OrdersByDeliveryDateAndFilterRequest) (*gen.OrdersWithRevisionResponse, error) {
	return rest.Call[gen.OrdersWithRevisionResponse](ctx, s.c, getDeliveriesByDeliveryDateAndSourceKeyAndFilter, req)
}

// GetDeliveriesByDeliveryDateAndStatus retrieve list of orders by statuses and dates.
func (s *Service) GetDeliveriesByDeliveryDateAndStatus(ctx context.Context, req gen.OrdersByDeliveryDateAndStatusRequest) (*gen.OrdersWithRevisionResponse, error) {
	return rest.Call[gen.OrdersWithRevisionResponse](ctx, s.c, getDeliveriesByDeliveryDateAndStatus, req)
}

// GetDeliveriesByID retrieve orders by IDs.
func (s *Service) GetDeliveriesByID(ctx context.Context, req gen.OrdersByIDRequest) (*gen.OrdersResponse, error) {
	return rest.Call[gen.OrdersResponse](ctx, s.c, getDeliveriesByID, req)
}

// GetDeliveriesByRevision retrieve list of orders changed from the time revision was passed.
func (s *Service) GetDeliveriesByRevision(ctx context.Context, req gen.OrdersByRevisionRequest) (*gen.OrdersWithRevisionResponse, error) {
	return rest.Call[gen.OrdersWithRevisionResponse](ctx, s.c, getDeliveriesByRevision, req)
}

// GetDeliveriesDraftsByFilter retrieve order drafts list by parameters.
func (s *Service) GetDeliveriesDraftsByFilter(ctx context.Context, req gen.FilterDraftsRequest) (*gen.FilterDraftsResponse, error) {
	return rest.Call[gen.FilterDraftsResponse](ctx, s.c, getDeliveriesDraftsByFilter, req)
}

// GetDeliveriesDraftsByID retrieve order draft by ID.
func (s *Service) GetDeliveriesDraftsByID(ctx context.Context, req gen.GetDraftRequest) (*gen.GetDraftResponse, error) {
	return rest.Call[gen.GetDraftResponse](ctx, s.c, getDeliveriesDraftsByID, req)
}

// GetDeliveriesHistoryByDeliveryDateAndPhone retrieve list of history orders by telephone number, dates and revision.
func (s *Service) GetDeliveriesHistoryByDeliveryDateAndPhone(ctx context.Context, req gen.OrdersHistoryByDeliveryDateAndPhoneRequest) (*gen.OrdersWithRevisionResponse, error) {
	return rest.Call[gen.OrdersWithRevisionResponse](ctx, s.c, getDeliveriesHistoryByDeliveryDateAndPhone, req)
}

// GetOrderByID retrieve orders by IDs.
func (s *Service) GetOrderByID(ctx context.Context, req gen.GetTableOrdersByIDRequest) (*gen.TableOrdersResponse, error) {
	return rest.Call[gen.TableOrdersResponse](ctx, s.c, getOrderByID, req)
}

// GetOrderByTable retrieve orders by tables.
func (s *Service) GetOrderByTable(ctx context.Context, req gen.GetTableOrdersByTableRequest) (*gen.TableOrdersResponse, error) {
	return rest.Call[gen.TableOrdersResponse](ctx, s.c, getOrderByTable, req)
}

// GetReserveAvailableOrganizations returns all organizations of current account (determined by Authorization request header) for which banquet/reserve booking are available.
func (s *Service) GetReserveAvailableOrganizations(ctx context.Context, req gen.GetOrganizationsRequest) (*gen.GetOrganizationsResponse, error) {
	return rest.Call[gen.GetOrganizationsResponse](ctx, s.c, getReserveAvailableOrganizations, req)
}

// GetReserveAvailableRestaurantSections returns all restaurant sections of specified terminal groups, for which banquet/reserve booking are available.
func (s *Service) GetReserveAvailableRestaurantSections(ctx context.Context, req gen.GetRestaurantSectionsRequest) (*gen.GetRestaurantSectionsResponse, error) {
	return rest.Call[gen.GetRestaurantSectionsResponse](ctx, s.c, getReserveAvailableRestaurantSections, req)
}

// GetReserveAvailableTerminalGroups returns all terminal groups of specified organizations, for which banquet/reserve booking are available.
func (s *Service) GetReserveAvailableTerminalGroups(ctx context.Context, req gen.GetTerminalGroupsByOrganizationsRequest) (*gen.TerminalGroupsResponse, error) {
	return rest.Call[gen.TerminalGroupsResponse](ctx, s.c, getReserveAvailableTerminalGroups, req)
}

// GetReserveRestaurantSectionsWorkload returns all banquets/reserves for passed restaurant sections.
func (s *Service) GetReserveRestaurantSectionsWorkload(ctx context.Context, req gen.GetRestaurantSectionsWorkloadRequest) (*gen.GetRestaurantSectionsWorkloadResponse, error) {
	return rest.Call[gen.GetRestaurantSectionsWorkloadResponse](ctx, s.c, getReserveRestaurantSectionsWorkload, req)
}

// GetReserveStatusByID retrieve banquets/reserves statuses by IDs.
func (s *Service) GetReserveStatusByID(ctx context.Context, req gen.ReservesByIDRequest) (*gen.ReservesResponse, error) {
	return rest.Call[gen.ReservesResponse](ctx, s.c, getReserveStatusByID, req)
}

// InitOrderByPosOrder init orders, created on POS, by POS orders.
func (s *Service) InitOrderByPosOrder(ctx context.Context, req gen.InitTableOrderByPosOrderRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, initOrderByPosOrder, req)
}

// InitOrderByTable init orders, created on POS, by tables.
func (s *Service) InitOrderByTable(ctx context.Context, req gen.InitTableOrderRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, initOrderByTable, req)
}

// LockDeliveriesDrafts lock order draft.
func (s *Service) LockDeliveriesDrafts(ctx context.Context, req gen.LockOrUnlockDraftRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, lockDeliveriesDrafts, req)
}

// PrintDeliveryBill print delivery bill.
func (s *Service) PrintDeliveryBill(ctx context.Context, req gen.PrintDeliveryBillRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, printDeliveryBill, req)
}

// PrintOrderBill print bill.
func (s *Service) PrintOrderBill(ctx context.Context, req gen.PrintBillRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, printOrderBill, req)
}

// SaveDeliveriesDrafts update existing delivery order draft.
func (s *Service) SaveDeliveriesDrafts(ctx context.Context, req gen.SaveDraftRequest) (*gen.CreateOrSaveDraftResponse, error) {
	return rest.Call[gen.CreateOrSaveDraftResponse](ctx, s.c, saveDeliveriesDrafts, req)
}

// UnlockDeliveriesDrafts unlock order draft.
func (s *Service) UnlockDeliveriesDrafts(ctx context.Context, req gen.LockOrUnlockDraftRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, unlockDeliveriesDrafts, req)
}

// UpdateOrderDeliveryStatusDeliveries update delivery status.
func (s *Service) UpdateOrderDeliveryStatusDeliveries(ctx context.Context, req gen.UpdateDeliveryStatusRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, updateOrderDeliveryStatusDeliveries, req)
}

// UpdateOrderProblemDeliveries update order problem.
func (s *Service) UpdateOrderProblemDeliveries(ctx context.Context, req gen.UpdateOrderProblemRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, updateOrderProblemDeliveries, req)
}

// UpdateTrackingLinkDeliveries update tracking link of an order.
func (s *Service) UpdateTrackingLinkDeliveries(ctx context.Context, req gen.UpdateTrackingLinkRequest) error {
	return s.c.Post(ctx, updateTrackingLinkDeliveries, req, nil)
}
