package documents

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

const okResult = `<documentValidationResult><valid>true</valid><documentNumber>N-1</documentNumber></documentValidationResult>`

// byNumber's three parameters interlock: currentYear is mandatory, true forbids
// from/to, false requires both. Getting it wrong reads a different year.
func TestExportByNumberEnforcesTheCurrentYearRule(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-01-01")
	to, _ := rest.ParseDay("2026-12-31")

	tests := []struct {
		name    string
		req     ByNumberRequest
		wantErr string
		wantQ   map[string]string
	}{
		{
			name:    "no number",
			req:     ByNumberRequest{CurrentYear: boolp(true)},
			wantErr: "number is required",
		},
		{
			name:    "currentYear unset",
			req:     ByNumberRequest{Number: "N-1"},
			wantErr: "currentYear is required",
		},
		{
			name:    "currentYear true with a window",
			req:     ByNumberRequest{Number: "N-1", CurrentYear: boolp(true), From: &from, To: &to},
			wantErr: "must not carry from/to",
		},
		{
			name:    "currentYear false without a window",
			req:     ByNumberRequest{Number: "N-1", CurrentYear: boolp(false)},
			wantErr: "requires both from and to",
		},
		{
			name:  "currentYear true alone",
			req:   ByNumberRequest{Number: "N-1", CurrentYear: boolp(true)},
			wantQ: map[string]string{"number": "N-1", "currentYear": "true"},
		},
		{
			name: "currentYear false with a window",
			req:  ByNumberRequest{Number: "N-1", CurrentYear: boolp(false), From: &from, To: &to},
			// v1 export dates are ISO here, unlike the import side's dotted form.
			wantQ: map[string]string{"currentYear": "false", "from": "2026-01-01", "to": "2026-12-31"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotQ url.Values
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotQ = r.URL.Query()
				_, _ = w.Write([]byte(`<incomingInvoiceDtoes></incomingInvoiceDtoes>`))
			})
			_, err := s.ExportIncomingInvoiceByNumber(context.Background(), tt.req)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("export: %v", err)
			}
			for k, want := range tt.wantQ {
				if gotQ.Get(k) != want {
					t.Errorf("%s = %q, want %q", k, gotQ.Get(k), want)
				}
			}
		})
	}
}

// A rejected document arrives as HTTP 200 with <valid>false</valid>. Every
// import must read the body, not the status.
func TestImportsTreatValidFalseAsFailure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		call func(*Service) error
		want string
	}{
		{
			name: "incoming invoice", want: EndpointImportIncomingInvoice,
			call: func(s *Service) error {
				_, err := s.ImportIncomingInvoice(context.Background(), IncomingInvoiceDto{DefaultStore: "st-1"})
				return err
			},
		},
		{
			name: "production document", want: EndpointImportProductionDocument,
			call: func(s *Service) error {
				_, err := s.ImportProductionDocument(context.Background(), ProductionDocumentDto{StoreFrom: "a", StoreTo: "b"})
				return err
			},
		},
		{
			name: "sales document", want: EndpointImportSalesDocument,
			call: func(s *Service) error {
				_, err := s.ImportSalesDocument(context.Background(), SalesDocumentDto{})
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(http.StatusOK) // 200, and still a rejection
				_, _ = w.Write([]byte(`<documentValidationResult><valid>false</valid><errorMessage>Склад не указан</errorMessage></documentValidationResult>`))
			})
			err := tt.call(s)
			if err == nil {
				t.Fatal("HTTP 200 with <valid>false</valid> is a rejection, not a success")
			}
			if !strings.Contains(err.Error(), "Склад не указан") {
				t.Errorf("iiko's own message must reach the caller, got %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.want) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.want)
			}
		})
	}
}

// Store placement differs per document type and is the commonest source of 409s,
// so each type is checked before a request is spent.
func TestImportsCheckTheirOwnStorePlacementRule(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		call    func(*Service) error
		wantErr string
	}{
		{
			name: "incoming invoice needs a store somewhere",
			call: func(s *Service) error {
				_, err := s.ImportIncomingInvoice(context.Background(), IncomingInvoiceDto{})
				return err
			},
			wantErr: "defaultStore",
		},
		{
			name: "production document needs both ends",
			call: func(s *Service) error {
				_, err := s.ImportProductionDocument(context.Background(), ProductionDocumentDto{StoreFrom: "a"})
				return err
			},
			wantErr: "storeFrom and storeTo",
		},
		{
			name: "returned invoice needs the original invoice identity",
			call: func(s *Service) error {
				_, err := s.ImportReturnedInvoice(context.Background(), ReturnedInvoiceDto{DefaultStoreID: "st-1"})
				return err
			},
			wantErr: "incomingInvoiceNumber and incomingInvoiceDate",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
				t.Error("must not spend a request on a document iiko will reject")
				w.WriteHeader(http.StatusInternalServerError)
			})
			err := tt.call(s)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("want %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestUnprocessHitsItsOwnPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		call func(*Service) error
		want string
	}{
		{
			name: "incoming", want: EndpointUnprocessIncomingInvoice,
			call: func(s *Service) error { return s.UnprocessIncomingInvoice(context.Background(), "N-1") },
		},
		{
			name: "outgoing", want: EndpointUnprocessOutgoingInvoice,
			call: func(s *Service) error { return s.UnprocessOutgoingInvoice(context.Background(), "N-1") },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = w.Write([]byte(okResult))
			})
			if err := tt.call(s); err != nil {
				t.Fatalf("unprocess: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.want) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.want)
			}
		})
	}
}

// byNumber answers a bare ARRAY where byId answers a bare object.
func TestV2ByNumberDecodesABareArray(t *testing.T) {
	t.Parallel()
	var gotQ url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		_, _ = w.Write([]byte(`[{"id":"w-1","documentNumber":"W-1","status":"PROCESSED"}]`))
	})
	got, err := s.DocumentsByNumber(context.Background(), KindWriteoff, "W-1")
	if err != nil {
		t.Fatalf("byNumber: %v", err)
	}
	if len(got) != 1 || got[0].DocumentNumber != "W-1" {
		t.Fatalf("decoded %+v", got)
	}
	if gotQ.Get("documentNumber") != "W-1" {
		t.Errorf("documentNumber = %q", gotQ.Get("documentNumber"))
	}
}

func boolp(b bool) *bool { return &b }

// The remaining v1 surface: the path each hits and the window format it sends.
func TestV1DocumentSurfaceHitsItsPaths(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	to, _ := rest.ParseDay("2026-03-31")

	tests := []struct {
		name     string
		body     string
		call     func(*Service) error
		wantPath string
		wantWin  bool
	}{
		{
			name: "export incoming invoices", wantPath: EndpointExportIncomingInvoice, wantWin: true,
			body: `<incomingInvoiceDtoes><document><documentNumber>N-1</documentNumber></document></incomingInvoiceDtoes>`,
			call: func(s *Service) error {
				got, err := s.ExportIncomingInvoices(context.Background(), from, to, "sup-1", -1)
				if err == nil && len(got) != 1 {
					return fmt.Errorf("decoded %d documents, want 1", len(got))
				}
				return err
			},
		},
		{
			name: "export outgoing invoices", wantPath: EndpointExportOutgoingInvoice, wantWin: true,
			body: `<outgoingInvoiceDtoes><document><documentNumber>N-2</documentNumber></document></outgoingInvoiceDtoes>`,
			call: func(s *Service) error {
				got, err := s.ExportOutgoingInvoices(context.Background(), from, to, "")
				if err == nil && len(got) != 1 {
					return fmt.Errorf("decoded %d documents, want 1", len(got))
				}
				return err
			},
		},
		{
			name: "export outgoing by number", wantPath: EndpointExportOutgoingInvoiceByNumber,
			body: `<outgoingInvoiceDtoes></outgoingInvoiceDtoes>`,
			call: func(s *Service) error {
				_, err := s.ExportOutgoingInvoiceByNumber(context.Background(),
					ByNumberRequest{Number: "N-2", CurrentYear: boolp(true)})
				return err
			},
		},
		{
			name: "import outgoing invoice", wantPath: EndpointImportOutgoingInvoice,
			body: okResult,
			call: func(s *Service) error {
				_, err := s.ImportOutgoingInvoice(context.Background(), OutgoingInvoiceDto{DefaultStoreID: "st-1"})
				return err
			},
		},
		{
			name: "import inventory", wantPath: EndpointImportIncomingInventory,
			body: `<incomingInventoryValidationResult><valid>true</valid></incomingInventoryValidationResult>`,
			call: func(s *Service) error {
				_, err := s.ImportIncomingInventory(context.Background(), InventoryDocument{
					StoreID: "st-1", Items: []InventoryItem{{ProductID: "p-1", AmountContainer: 1}},
				})
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			var gotQ url.Values
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotQ = r.URL.Path, r.URL.Query()
				_, _ = w.Write([]byte(tt.body))
			})
			if err := tt.call(s); err != nil {
				t.Fatalf("call: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.wantPath) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.wantPath)
			}
			if tt.wantWin {
				// v1 document EXPORT takes ISO, unlike the v1 report family's dotted form.
				if gotQ.Get("from") != "2026-03-01" || gotQ.Get("to") != "2026-03-31" {
					t.Errorf("export window must be ISO, got from=%q to=%q", gotQ.Get("from"), gotQ.Get("to"))
				}
			}
		})
	}
}

// The inventory import persists where CheckInventory does not, so its rejection
// path has to be just as loud.
func TestImportInventoryRejectsAndValidates(t *testing.T) {
	t.Parallel()
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<incomingInventoryValidationResult><valid>false</valid><errorMessage>Склад не указан</errorMessage></incomingInventoryValidationResult>`))
	})
	_, err := s.ImportIncomingInventory(context.Background(), InventoryDocument{
		StoreID: "st-1", Items: []InventoryItem{{ProductID: "p-1"}},
	})
	if err == nil || !strings.Contains(err.Error(), "Склад не указан") {
		t.Errorf("a rejected inventory import must surface iiko's message, got %v", err)
	}
	if _, err := s.ImportIncomingInventory(context.Background(), InventoryDocument{StoreID: "st-1"}); err == nil {
		t.Error("an inventory with no counted items must be refused before the request")
	}
}
