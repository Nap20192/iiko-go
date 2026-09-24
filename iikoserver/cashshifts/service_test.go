package cashshifts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
	"github.com/Nap20192/iiko-go/iikoserver/rest/resttest"
)

// newFake wires a Service to a stub iiko; the transport is returned too, for the
// tests that close the session.
func newFake(t *testing.T, h http.HandlerFunc) (*Service, *rest.Client) {
	t.Helper()
	c := resttest.NewFake(t, h)
	return New(c), c
}

// The docs misspell three field names in their own tables while every worked
// example spells them correctly. A client that trusts only one spelling silently
// reports a zero where a real figure was, so both must decode.
func TestCashShiftSessionAcceptsBothDocumentedSpellings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		body       string
		wantStatus string
		wantCredit float64
		wantMgr    string
	}{
		{
			name:       "example spellings",
			body:       `{"sessionStatus":"UNACCEPTED","salesCredit":10.5,"managerId":"m-1"}`,
			wantStatus: "UNACCEPTED", wantCredit: 10.5, wantMgr: "m-1",
		},
		{
			name:       "field-table spellings",
			body:       `{"sessionStaus":"HASWARNINGS","salesCerdit":7.25,"manager":"m-2"}`,
			wantStatus: "HASWARNINGS", wantCredit: 7.25, wantMgr: "m-2",
		},
		{
			name:       "both present, correct spelling wins",
			body:       `{"sessionStatus":"ACCEPTED","sessionStaus":"UNACCEPTED","salesCredit":1,"salesCerdit":2,"managerId":"m-a","manager":"m-b"}`,
			wantStatus: "ACCEPTED", wantCredit: 1, wantMgr: "m-a",
		},
		{
			name:       "zero salesCerdit is not mistaken for absent",
			body:       `{"salesCerdit":0}`,
			wantCredit: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got CashShiftSession
			if err := json.Unmarshal([]byte(tt.body), &got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got.SessionStatus != tt.wantStatus {
				t.Errorf("sessionStatus = %q, want %q", got.SessionStatus, tt.wantStatus)
			}
			if got.SalesCredit != tt.wantCredit {
				t.Errorf("salesCredit = %v, want %v", got.SalesCredit, tt.wantCredit)
			}
			if got.ManagerID != tt.wantMgr {
				t.Errorf("managerId = %q, want %q", got.ManagerID, tt.wantMgr)
			}
		})
	}
}

func TestCashShiftPaymentsAcceptsBothDocumentedSpellings(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, body string }{
		{name: "example spelling", body: `{"payOutRecords":[{"actualSum":5}]}`},
		{name: "field-table spelling", body: `{"payOutsRecords":[{"actualSum":5}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got CashShiftPayments
			if err := json.Unmarshal([]byte(tt.body), &got); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(got.PayOutRecords) != 1 || got.PayOutRecords[0].ActualSum != 5 {
				t.Errorf("pay-out records lost: %+v", got.PayOutRecords)
			}
		})
	}
}

func TestTransactionInfoAcceptsBothCauseEventSpellings(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`{"causeEventId":"e-1"}`, `{"causeEvenId":"e-1"}`} {
		var got TransactionInfo
		if err := json.Unmarshal([]byte(body), &got); err != nil {
			t.Fatalf("decode %s: %v", body, err)
		}
		if got.CauseEventID != "e-1" {
			t.Errorf("%s lost causeEventId, got %q", body, got.CauseEventID)
		}
	}
}

func TestListCashShiftsRequiresAStatus(t *testing.T) {
	t.Parallel()
	s, c := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not reach iiko: iiko rejects an empty status, so we reject it first")
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer c.Close(context.Background())
	day := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if _, err := s.ListCashShifts(context.Background(), day, day, "  ", "", ""); err == nil ||
		!strings.Contains(err.Error(), "status is required") {
		t.Errorf("want a status-required error, got %v", err)
	}
}

func TestListCashShiftsSendsInclusiveDayBoundsAndFilters(t *testing.T) {
	t.Parallel()
	var gotQuery url.Values
	var gotPath string
	s, c := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery, gotPath = r.URL.Query(), r.URL.Path
		_, _ = w.Write([]byte(`[{"id":"s-1","sessionNumber":7,"sessionStatus":"CLOSED"}]`))
	})
	defer c.Close(context.Background())

	from := time.Date(2026, 3, 1, 13, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 22, 0, 0, 0, time.UTC)
	got, err := s.ListCashShifts(context.Background(), from, to, ShiftAny, "dep-1", "grp-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].ID != "s-1" {
		t.Fatalf("want one session, got %+v", got)
	}
	if !strings.HasSuffix(gotPath, EndpointCashShiftsList) {
		t.Errorf("wrong path %q", gotPath)
	}
	for k, want := range map[string]string{
		"openDateFrom": "2026-03-01",
		"openDateTo":   "2026-03-31",
		"status":       "ANY",
		"departmentId": "dep-1",
		"groupId":      "grp-1",
	} {
		if gotQuery.Get(k) != want {
			t.Errorf("%s = %q, want %q", k, gotQuery.Get(k), want)
		}
	}
}

func TestCashShiftByIDAndPaymentsRequireASession(t *testing.T) {
	t.Parallel()
	s, c := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not reach iiko without a session id")
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer c.Close(context.Background())
	if _, err := s.CashShiftByID(context.Background(), " "); err == nil {
		t.Error("byId must require a session id")
	}
	if _, err := s.CashShiftPaymentsByID(context.Background(), " ", false); err == nil {
		t.Error("payments must require a session id")
	}
}

// byId answers a bare object where list answers a bare array; both have to work
// through the same decode path.
func TestCashShiftByIDDecodesABareObject(t *testing.T) {
	t.Parallel()
	var gotPath string
	s, c := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"s-9","sessionNumber":3,"cashRemain":12.5,"sessionStaus":"ACCEPTED"}`))
	})
	defer c.Close(context.Background())
	got, err := s.CashShiftByID(context.Background(), "s-9")
	if err != nil {
		t.Fatalf("byId: %v", err)
	}
	if got.ID != "s-9" || got.CashRemain != 12.5 || got.SessionStatus != "ACCEPTED" {
		t.Errorf("decoded wrong: %+v", got)
	}
	if !strings.Contains(gotPath, "/byId/s-9") {
		t.Errorf("session id must be a path segment, got %q", gotPath)
	}
}

func TestCashShiftPaymentsSendsHideAcceptedExplicitly(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		hide bool
		want string
	}{
		{name: "hide accepted", hide: true, want: "true"},
		{name: "keep accepted", hide: false, want: "false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got string
			s, c := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				got = r.URL.Query().Get("hideAccepted")
				_, _ = w.Write([]byte(`{"sessionId":"s-1"}`))
			})
			defer c.Close(context.Background())
			if _, err := s.CashShiftPaymentsByID(context.Background(), "s-1", tt.hide); err != nil {
				t.Fatalf("payments: %v", err)
			}
			if got != tt.want {
				t.Errorf("hideAccepted = %q, want %q sent explicitly", got, tt.want)
			}
		})
	}
}
