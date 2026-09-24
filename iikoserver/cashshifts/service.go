package cashshifts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the cash shifts and their payments half of the iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// Cash shifts, v2 JSON, iiko 5.4. Field names come from research/dtos.yaml
// (CashShiftSessionDto, CashShiftPaymentsDto, PaymentRecordDto).
//
// Three fields are misspelled in the docs' own field tables and spelled
// correctly in every worked example. Since the examples are captured responses
// they are the likelier truth, so the correct spelling is primary and the
// table's spelling is accepted as a fallback. See CashShiftSession.UnmarshalJSON.

// ShiftStatus values accepted by the list filter. The filter is mandatory
// upstream and rejects an empty value, so callers must send one.
//
// ANY, OPEN and CLOSED are filter-only. A session's own sessionStatus is one of
// ACCEPTED, UNACCEPTED or HASWARNINGS.
const (
	ShiftAny         = "ANY"
	ShiftOpen        = "OPEN"
	ShiftClosed      = "CLOSED"
	ShiftAccepted    = "ACCEPTED"
	ShiftUnaccepted  = "UNACCEPTED"
	ShiftHasWarnings = "HASWARNINGS"
)

// ShiftStatuses is the documented filter vocabulary, in the docs' order.
func ShiftStatuses() []any {
	return []any{ShiftAny, ShiftOpen, ShiftClosed, ShiftAccepted, ShiftUnaccepted, ShiftHasWarnings}
}

// CashShiftSession is one till session.
type CashShiftSession struct {
	ID            string `json:"id"`
	SessionNumber int    `json:"sessionNumber"`
	FiscalNumber  int    `json:"fiscalNumber"`
	CashRegNumber int    `json:"cashRegNumber"`
	CashRegSerial string `json:"cashRegSerial"`

	OpenDate   string `json:"openDate"`
	CloseDate  string `json:"closeDate"`
	AcceptDate string `json:"acceptDate"`

	ManagerID       string `json:"managerId"`
	ResponsibleUser string `json:"responsibleUser"`

	SessionStartCash  float64 `json:"sessionStartCash"`
	PayOrders         float64 `json:"payOrders"`
	SumWriteoffOrders float64 `json:"sumWriteoffOrders"`
	SalesCash         float64 `json:"salesCash"`
	SalesCredit       float64 `json:"salesCredit"`
	SalesCard         float64 `json:"salesCard"`
	PayIn             float64 `json:"payIn"`
	PayOut            float64 `json:"payOut"`
	PayIncome         float64 `json:"payIncome"`
	CashRemain        float64 `json:"cashRemain"`
	CashDiff          float64 `json:"cashDiff"`

	SessionStatus string `json:"sessionStatus"`
	Conception    string `json:"conception"`
	PointOfSale   string `json:"pointOfSale"`
}

// UnmarshalJSON accepts both documented spellings of the three fields the docs
// disagree with themselves about: sessionStatus/sessionStaus,
// salesCredit/salesCerdit and managerId/manager. The correctly spelled form
// wins when both are present.
func (s *CashShiftSession) UnmarshalJSON(b []byte) error {
	type plain CashShiftSession // no method set, so no recursion
	var v struct {
		plain
		SessionStaus string   `json:"sessionStaus"`
		SalesCerdit  *float64 `json:"salesCerdit"`
		Manager      string   `json:"manager"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*s = CashShiftSession(v.plain)
	if s.SessionStatus == "" {
		s.SessionStatus = v.SessionStaus
	}
	if s.ManagerID == "" {
		s.ManagerID = v.Manager
	}
	if s.SalesCredit == 0 && v.SalesCerdit != nil {
		s.SalesCredit = *v.SalesCerdit
	}
	return nil
}

// CashShiftPayments is the per-session payment breakdown.
type CashShiftPayments struct {
	SessionID    string `json:"sessionId"`
	OperationDay string `json:"operationDay"`

	CashlessRecords []PaymentRecord `json:"cashlessRecords"`
	PayInRecords    []PaymentRecord `json:"payInRecords"`
	PayOutRecords   []PaymentRecord `json:"payOutRecords"`
}

// UnmarshalJSON accepts payOutRecords and the field table's payOutsRecords.
func (p *CashShiftPayments) UnmarshalJSON(b []byte) error {
	type plain CashShiftPayments
	var v struct {
		plain
		PayOutsRecords []PaymentRecord `json:"payOutsRecords"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = CashShiftPayments(v.plain)
	if len(p.PayOutRecords) == 0 {
		p.PayOutRecords = v.PayOutsRecords
	}
	return nil
}

// PaymentRecord is one cashless payment, pay-in or pay-out.
//
// ActualSum differs from OriginalSum when a manager edited the figure while
// accepting the shift; iiko permits that only for pay-outs.
type PaymentRecord struct {
	Info            TransactionInfo `json:"info"`
	ActualSum       float64         `json:"actualSum"`
	OriginalSum     float64         `json:"originalSum"`
	PaymentTypeID   string          `json:"paymentTypeId"`
	PayAgentID      string          `json:"payAgentId"`
	EditableComment string          `json:"editableComment"`
	Status          string          `json:"status"`
}

// TransactionInfo is the accounting context of a payment record.
type TransactionInfo struct {
	ID           string  `json:"id"`
	Date         string  `json:"date"`
	Sum          float64 `json:"sum"`
	CauseEventID string  `json:"causeEventId"`
	Comment      string  `json:"comment"`
}

// UnmarshalJSON accepts causeEventId and the field table's causeEvenId.
func (t *TransactionInfo) UnmarshalJSON(b []byte) error {
	type plain TransactionInfo
	var v struct {
		plain
		CauseEvenID string `json:"causeEvenId"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*t = TransactionInfo(v.plain)
	if t.CauseEventID == "" {
		t.CauseEventID = v.CauseEvenID
	}
	return nil
}

// ListCashShifts returns till sessions opened in [from, to]; both bounds are
// inclusive upstream.
//
// status is mandatory: iiko rejects an empty value, so callers pass ShiftAny
// rather than omitting it.
func (s *Service) ListCashShifts(ctx context.Context, from, to time.Time, status, departmentID, groupID string) ([]CashShiftSession, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return nil, fmt.Errorf("status is required: iiko rejects an empty cash-shift status; pass %s for no filtering", ShiftAny)
	}
	q := url.Values{
		"openDateFrom": {from.Format(rest.QueryV2)},
		"openDateTo":   {to.Format(rest.QueryV2)},
		"status":       {status},
	}
	if departmentID != "" {
		q.Set("departmentId", departmentID)
	}
	if groupID != "" {
		q.Set("groupId", groupID)
	}
	var out []CashShiftSession
	if _, err := s.rest.GetV2(ctx, EndpointCashShiftsList, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CashShiftByID returns one session. The response is a bare object, unlike
// list's bare array.
func (s *Service) CashShiftByID(ctx context.Context, sessionID string) (*CashShiftSession, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("sessionId is required")
	}
	var out CashShiftSession
	if _, err := s.rest.GetV2(ctx, EndpointCashShiftByID+url.PathEscape(sessionID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CashShiftPaymentsByID returns the payment breakdown for one session.
// hideAccepted drops records already accepted by a manager.
func (s *Service) CashShiftPaymentsByID(ctx context.Context, sessionID string, hideAccepted bool) (*CashShiftPayments, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("sessionId is required")
	}
	q := url.Values{"hideAccepted": {rest.BoolStr(hideAccepted)}}
	var out CashShiftPayments
	if _, err := s.rest.GetV2(ctx, EndpointCashShiftPayments+url.PathEscape(sessionID), q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
