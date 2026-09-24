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

// SessionRef identifies the shift an acceptance document belongs to.
type SessionRef struct {
	ID      string `json:"id"`
	GroupID string `json:"groupId"`
	Number  int    `json:"sessionNumber"`
}

// ClosedSessionItem is one line a manager accepts or corrects.
//
// SumReal is the figure the manager counted; iiko honours an edit to it only on
// pay-outs, and silently keeps its own figure everywhere else.
type ClosedSessionItem struct {
	Type        string  `json:"type"` // CARD | CREDIT | PAYIN | PAYOUT
	Status      string  `json:"status"`
	SumExpected float64 `json:"sumExpected"`
	SumReal     float64 `json:"sumReal"`
	Comment     string  `json:"comment,omitempty"`
}

// ClosedSessionDocument is a shift's acceptance document.
type ClosedSessionDocument struct {
	ID      string              `json:"id"`
	Session SessionRef          `json:"session"`
	Items   []ClosedSessionItem `json:"items"`
}

// ClosedSessionDocument returns a shift's acceptance document, CREATING it if
// the shift does not have one yet. It is a GET that writes: calling it to look
// at a shift leaves a document behind.
func (s *Service) ClosedSessionDocument(ctx context.Context, sessionID string) (*ClosedSessionDocument, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("sessionId is required: this call creates an acceptance document when none exists")
	}
	path := strings.Replace(EndpointClosedSessionDocument, "{id}", url.PathEscape(sessionID), 1)
	var out ClosedSessionDocument
	if _, err := s.rest.GetV2(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AcceptShift saves a manager's acceptance of a shift.
//
// An edited SumReal is honoured only on pay-outs, so an edit anywhere else is
// refused here rather than sent and silently dropped.
func (s *Service) AcceptShift(ctx context.Context, doc ClosedSessionDocument) error {
	if strings.TrimSpace(doc.ID) == "" {
		return fmt.Errorf("document id is required")
	}
	for i, it := range doc.Items {
		if it.Type == "PAYOUT" {
			continue
		}
		if it.SumReal != 0 && it.SumReal != it.SumExpected {
			return fmt.Errorf("items[%d] (%s): sumReal is editable only for pay-outs; iiko keeps its own figure for everything else and the correction would be silently lost", i, it.Type)
		}
	}
	var out ClosedSessionDocument
	_, err := s.rest.PostV2(ctx, EndpointCashShiftSave, nil, doc, &out)
	return err
}

// PayOutRequest authorises a cash withdrawal from a till.
type PayOutRequest struct {
	SessionID  string  `json:"sessionId"`
	Sum        float64 `json:"sum"`
	AccountID  string  `json:"accountId,omitempty"`
	TypeID     string  `json:"payInOutTypeId,omitempty"`
	Comment    string  `json:"comment,omitempty"`
	EmployeeID string  `json:"employeeId,omitempty"`
}

// AddPayOutResult is what iiko returns after authorising a pay-out.
type AddPayOutResult struct {
	PayOutSettings map[string]any `json:"payOutSettings"`
}

// AddPayOut authorises a cash pay-out. Requires the F_APIO right.
//
// The request goes out as JSON with an explicit charset, which iiko needs for a
// Cyrillic comment; PostV2 sets it.
func (s *Service) AddPayOut(ctx context.Context, req PayOutRequest) (*AddPayOutResult, error) {
	if strings.TrimSpace(req.SessionID) == "" {
		return nil, fmt.Errorf("sessionId is required")
	}
	if req.Sum <= 0 {
		return nil, fmt.Errorf("sum must be positive")
	}
	var out AddPayOutResult
	if _, err := s.rest.PostV2(ctx, EndpointAddPayOut, nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PayInOutType is a reason code for a till pay-in or pay-out.
type PayInOutType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
}

// PayInOutTypes lists the reason codes. Requires the B_APIO right.
func (s *Service) PayInOutTypes(ctx context.Context, includeDeleted bool) ([]PayInOutType, error) {
	q := url.Values{"includeDeleted": {rest.BoolStr(includeDeleted)}}
	var out []PayInOutType
	if _, err := s.rest.GetV2(ctx, EndpointPayInOutTypes, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Payroll is one employee's accrued pay for a period.
type Payroll struct {
	ID         string  `json:"id"`
	EmployeeID string  `json:"employeeId"`
	Sum        float64 `json:"sum"`
}

// UnmarshalJSON accepts the field table's payrollId alongside the example's id.
func (p *Payroll) UnmarshalJSON(b []byte) error {
	type plain Payroll
	var v struct {
		plain
		PayrollID string `json:"payrollId"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*p = Payroll(v.plain)
	if p.ID == "" {
		p.ID = v.PayrollID
	}
	return nil
}

// Payrolls lists accrued pay. dateFrom and dateTo are both inclusive.
func (s *Service) Payrolls(ctx context.Context, from, to time.Time, departmentID string, includeDeleted bool) ([]Payroll, error) {
	q := url.Values{
		"dateFrom":       {from.Format(rest.QueryV2)},
		"dateTo":         {to.Format(rest.QueryV2)},
		"includeDeleted": {rest.BoolStr(includeDeleted)},
	}
	if departmentID != "" {
		q.Set("department", departmentID)
	}
	var out []Payroll
	if _, err := s.rest.GetV2(ctx, EndpointPayrolls, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}
