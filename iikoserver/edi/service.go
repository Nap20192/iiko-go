// Package edi covers EDI order exchange: reading orders a supplier sent,
// creating and editing them, moving them through draft → readyToSend →
// registered, and answering them with confirmations and invoices.
package edi

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the EDI half of the iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// Acknowledgement kinds for orders/ack: confirm receipt of the order, or of
// iikoServer's cancellation of it.
const (
	AckOriginal = "original"
	AckCanceled = "canceled"
)

// SellerFilter narrows orders/bySeller. One of GLN or INN is required: without
// a seller the call is not a wildcard, it is a filter error.
type SellerFilter struct {
	GLN  string
	INN  string
	KPP  string
	Name string
}

// documentIdentifiers is the request body register, send and unregister share:
// a bare array of the {number, date} keys that identify a document.
type documentIdentifiers struct {
	XMLName    xml.Name                `xml:"documentIdentifierDtoes"`
	Identifier []DocumentIdentifierDto `xml:"documentIdentifier"`
}

// orderList is the <orderDtoes> array those three calls answer with.
type orderList struct {
	XMLName xml.Name   `xml:"orderDtoes"`
	Order   []OrderDto `xml:"order"`
}

// messageList is the <ediMessageDtoes> array bySeller answers with.
type messageList struct {
	XMLName xml.Name        `xml:"ediMessageDtoes"`
	Message []EdiMessageDto `xml:"ediMessage"`
}

// endpoint fills an EDI system's GUID into a path template.
func endpoint(template, ediSystem string) string {
	return fmt.Sprintf(template, url.PathEscape(ediSystem))
}

// SendInvoice confirms shipment and submits the invoice. The docs give no
// response schema, so nothing is decoded.
func (s *Service) SendInvoice(ctx context.Context, ediSystem string, msg EdiMessageDto) error {
	if err := requireSystem(ediSystem); err != nil {
		return err
	}
	return s.rest.PutXML(ctx, endpoint(EndpointInvoice, ediSystem), nil, msg, nil)
}

// AckOrder acknowledges receipt of an order, or of its cancellation. number and
// date are the document's identity key. An unacknowledged order keeps coming
// back from OrdersBySeller, so status defaults to AckOriginal rather than empty.
func (s *Service) AckOrder(ctx context.Context, ediSystem, number string, date time.Time, status string) error {
	if err := requireSystem(ediSystem); err != nil {
		return err
	}
	if strings.TrimSpace(number) == "" || date.IsZero() {
		return fmt.Errorf("number and date identify the order and are both required")
	}
	switch status = strings.TrimSpace(status); status {
	case "":
		status = AckOriginal
	case AckOriginal, AckCanceled:
	default:
		return fmt.Errorf("status %q is not documented: use %s or %s", status, AckOriginal, AckCanceled)
	}
	q := url.Values{"number": {number}, "date": {date.Format(rest.QueryV2)}, "status": {status}}
	// The only call here with no body at all: orders/ack is entirely query
	// string, and PutXML has nothing to marshal.
	_, err := s.rest.Do(ctx, http.MethodPut, endpoint(EndpointOrdersAck, ediSystem), q, nil, "")
	return err
}

// OrdersBySeller returns the EDI messages waiting for a seller, including
// orders iiko has since cancelled that were already acknowledged.
func (s *Service) OrdersBySeller(ctx context.Context, ediSystem string, f SellerFilter) ([]EdiMessageDto, error) {
	if err := requireSystem(ediSystem); err != nil {
		return nil, err
	}
	if strings.TrimSpace(f.GLN) == "" && strings.TrimSpace(f.INN) == "" {
		return nil, fmt.Errorf("one of gln or inn is required to identify the seller")
	}
	q := url.Values{}
	for k, v := range map[string]string{"gln": f.GLN, "inn": f.INN, "kpp": f.KPP, "name": f.Name} {
		if v != "" {
			q.Set(k, v)
		}
	}
	var out messageList
	if err := s.rest.GetXML(ctx, endpoint(EndpointOrdersBySeller, ediSystem), q, &out); err != nil {
		return nil, err
	}
	return out.Message, nil
}

// ListOrders returns every order in the window regardless of processing status.
// revisionFrom is exclusive; -1 asks for everything.
func (s *Service) ListOrders(ctx context.Context, ediSystem string, from, to time.Time,
	revisionFrom int64,
) ([]OrderDto, error) {
	if err := requireSystem(ediSystem); err != nil {
		return nil, err
	}
	q := url.Values{"revisionFrom": {strconv.FormatInt(revisionFrom, 10)}}
	if !from.IsZero() {
		q.Set("from", from.Format(rest.QueryV2))
	}
	if !to.IsZero() {
		q.Set("to", to.Format(rest.QueryV2))
	}
	var out orderList
	if err := s.rest.GetXML(ctx, endpoint(EndpointOrdersList, ediSystem), q, &out); err != nil {
		return nil, err
	}
	return out.Order, nil
}

// CreateOrder creates a draft order. It changes production data.
//
// Omit the number and iiko assigns one; supplying a number that already exists
// fails. The order always starts as a draft, so processingStatus is stripped.
// The result's warnings are the real payload — an order can be created with
// products that are not on the supplier's price list.
func (s *Service) CreateOrder(ctx context.Context, ediSystem string, order OrderDto) (*OrderCreationOrUpdateResultDto, error) {
	if err := requireSystem(ediSystem); err != nil {
		return nil, err
	}
	order.StripReadOnly()
	var out OrderCreationOrUpdateResultDto
	if err := s.rest.PostXML(ctx, endpoint(EndpointOrdersCreate, ediSystem), nil, order, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateOrder edits a draft in place. It changes production data.
//
// number and date are the identity key and cannot be changed: an order missing
// either addresses no document rather than creating one.
func (s *Service) UpdateOrder(ctx context.Context, ediSystem string, order OrderDto) (*OrderCreationOrUpdateResultDto, error) {
	if err := requireSystem(ediSystem); err != nil {
		return nil, err
	}
	if strings.TrimSpace(order.Number) == "" || strings.TrimSpace(order.Date) == "" {
		return nil, fmt.Errorf("number and date identify the order and cannot be omitted or modified")
	}
	order.StripReadOnly()
	var out OrderCreationOrUpdateResultDto
	if err := s.rest.PutXML(ctx, endpoint(EndpointOrdersUpdate, ediSystem), nil, order, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RegisterOrders posts drafts, moving them to processingStatus=registered.
func (s *Service) RegisterOrders(ctx context.Context, ediSystem string, keys []DocumentIdentifierDto) ([]OrderDto, error) {
	return s.moveOrders(ctx, EndpointOrdersRegister, ediSystem, keys)
}

// SendOrders queues orders for sending, moving them to readyToSend — the status
// that makes them visible to the EDI participant over OrdersBySeller.
func (s *Service) SendOrders(ctx context.Context, ediSystem string, keys []DocumentIdentifierDto) ([]OrderDto, error) {
	return s.moveOrders(ctx, EndpointOrdersSend, ediSystem, keys)
}

// UnregisterOrders un-posts orders, moving them back to draft.
func (s *Service) UnregisterOrders(ctx context.Context, ediSystem string, keys []DocumentIdentifierDto) ([]OrderDto, error) {
	return s.moveOrders(ctx, EndpointOrdersUnregister, ediSystem, keys)
}

// moveOrders is the three status transitions, which differ only in path.
func (s *Service) moveOrders(ctx context.Context, template, ediSystem string, keys []DocumentIdentifierDto) ([]OrderDto, error) {
	if err := requireSystem(ediSystem); err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("at least one {number, date} identifier is required")
	}
	var out orderList
	body := documentIdentifiers{Identifier: keys}
	if err := s.rest.PutXML(ctx, endpoint(template, ediSystem), nil, body, &out); err != nil {
		return nil, err
	}
	return out.Order, nil
}

// RespondToOrder answers an order line by line: confirm a line by resending it
// with its confirmedQuantity, cancel one with CancelLine, add one by sending a
// fresh lineNumber with no orderLineNumber. Amending a confirmed line means
// cancelling it and adding a replacement — there is no edit.
func (s *Service) RespondToOrder(ctx context.Context, ediSystem string, msg EdiMessageDto) (*EdiMessageDto, error) {
	if err := requireSystem(ediSystem); err != nil {
		return nil, err
	}
	var out EdiMessageDto
	if err := s.rest.PutXML(ctx, endpoint(EndpointResponse, ediSystem), nil, msg, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelLine returns the line zeroed out, which is how a confirmed line is
// cancelled. Container quantity goes too: zeroing only the units cancels the
// weight while still confirming the boxes.
func CancelLine(item OrderResponseItemDto) OrderResponseItemDto {
	item.ConfirmedQuantity.Quantity = 0
	item.ConfirmedQuantity.ContainerQuantity = 0
	return item
}

func requireSystem(ediSystem string) error {
	if strings.TrimSpace(ediSystem) == "" {
		return fmt.Errorf("ediSystem GUID is required: find it in iikoOffice under Обмен данными → Системы EDI")
	}
	return nil
}
