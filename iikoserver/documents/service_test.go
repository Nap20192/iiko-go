package documents

import (
	"context"
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

func TestInventoryDateUsesTheDocumentLayout(t *testing.T) {
	t.Parallel()
	got := InventoryDate(time.Date(2026, 7, 3, 0, 26, 0, 0, time.UTC))
	if got != "2026-07-03T00:26:00" {
		t.Errorf("inventory dateIncoming must be yyyy-MM-ddTHH:mm:ss, got %q", got)
	}
}

func TestCheckInventoryRejectsUnsendableDocuments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		doc     InventoryDocument
		wantSub string
	}{
		{
			name:    "no items",
			doc:     InventoryDocument{StoreID: "s1"},
			wantSub: "at least one counted item",
		},
		{
			name:    "no store placement",
			doc:     InventoryDocument{Items: []InventoryItem{{ProductID: "p1", AmountContainer: 1}}},
			wantSub: "storeId or storeCode is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, c := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
				t.Error("must not reach iiko: the document is unsendable and wastes a serialized request")
				w.WriteHeader(http.StatusInternalServerError)
			})
			defer c.Close(context.Background())
			if _, err := s.CheckInventory(context.Background(), tt.doc); err == nil ||
				!strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("want error containing %q, got %v", tt.wantSub, err)
			}
		})
	}
}

// The dry run's whole value is the surplus/shortage table, and iiko reports a
// rejected document as HTTP 200 with <valid>false</valid>. Trusting the status
// code would turn a rejection into a confident empty report.
func TestCheckInventoryTreatsValidFalseAsFailure(t *testing.T) {
	t.Parallel()
	const okBody = `<incomingInventoryValidationResult>
  <valid>true</valid><warning>false</warning>
  <documentNumber>Imv1</documentNumber>
  <store><id>st-1</id><code>1</code><name>Main storage</name></store>
  <date>2026-07-03T00:26:00+03:00</date>
  <items><item>
    <product><id>p-1</id><code>00001</code><name>Milk</name></product>
    <expectedAmount>13.600000000</expectedAmount><expectedSum>535.37</expectedSum>
    <actualAmount>29.450</actualAmount>
    <differenceAmount>0</differenceAmount><differenceSum>0</differenceSum>
  </item></items>
</incomingInventoryValidationResult>`

	tests := []struct {
		name    string
		status  int
		body    string
		wantErr string
		wantRow string
	}{
		{name: "valid document returns the breakdown", status: 200, body: okBody, wantRow: "Milk"},
		{
			name:   "XSD root element name is also accepted",
			status: 200,
			body:   strings.ReplaceAll(okBody, "incomingInventoryValidationResult", "document"),
			// The XSD calls the root <document>; the worked example calls it
			// <incomingInventoryValidationResult>. Both must decode.
			wantRow: "Milk",
		},
		{
			name:    "valid false is an error even on HTTP 200",
			status:  200,
			body:    `<incomingInventoryValidationResult><valid>false</valid><errorMessage>Склад не указан</errorMessage></incomingInventoryValidationResult>`,
			wantErr: "Склад не указан",
		},
		{
			name:    "additionalInfo is appended to the message",
			status:  200,
			body:    `<incomingInventoryValidationResult><valid>false</valid><errorMessage>Ошибка</errorMessage><additionalInfo>строка 2</additionalInfo></incomingInventoryValidationResult>`,
			wantErr: "строка 2",
		},
		{
			name:    "valid false without a message still fails loudly",
			status:  200,
			body:    `<incomingInventoryValidationResult><valid>false</valid></incomingInventoryValidationResult>`,
			wantErr: "without a message",
		},
		{
			name:    "409 business error is passed through verbatim",
			status:  409,
			body:    "Инвентаризация за этот день уже проведена",
			wantErr: "уже проведена",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotBody, gotCT, gotPath string
			s, c := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				b := make([]byte, r.ContentLength)
				_, _ = r.Body.Read(b)
				gotBody, gotCT, gotPath = string(b), r.Header.Get("Content-Type"), r.URL.Path
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})
			defer c.Close(context.Background())

			doc := InventoryDocument{
				StoreID: "st-1",
				Status:  "NEW",
				Items:   []InventoryItem{{ProductID: "p-1", AmountContainer: 29.45}},
			}
			res, err := s.CheckInventory(context.Background(), doc)

			if !strings.HasSuffix(gotPath, EndpointInventoryCheck) {
				t.Errorf("wrong path %q, want suffix %q", gotPath, EndpointInventoryCheck)
			}
			if !strings.Contains(gotCT, "xml") {
				t.Errorf("v1 documents are XML in and XML out, got Content-Type %q", gotCT)
			}
			if !strings.Contains(gotBody, "<document>") {
				t.Errorf("request root element must be <document>, got %q", gotBody)
			}

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got result %+v", tt.wantErr, res)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q must carry iiko's own message %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(res.Items) != 1 || res.Items[0].Product.Name != tt.wantRow {
				t.Fatalf("want one row for %q, got %+v", tt.wantRow, res.Items)
			}
			if res.Items[0].ExpectedAmount != 13.6 || res.Items[0].ActualAmount != 29.45 {
				t.Errorf("expected/actual amounts decoded wrong: %+v", res.Items[0])
			}
		})
	}
}

func TestDocumentEndpointsRejectAnUnknownKind(t *testing.T) {
	t.Parallel()
	s, c := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not reach iiko for an unknown document kind")
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer c.Close(context.Background())
	day := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if _, err := s.ListDocuments(context.Background(), DocumentKind("salesDocument"), day, day, ""); err == nil ||
		!strings.Contains(err.Error(), "unknown document kind") {
		t.Errorf("want an unknown-kind error, got %v", err)
	}
	if _, err := s.DocumentByID(context.Background(), DocumentKind("nope"), "d-1"); err == nil {
		t.Error("byId must reject an unknown kind too")
	}
	if _, err := s.DocumentByID(context.Background(), KindWriteoff, " "); err == nil {
		t.Error("byId must require an id")
	}
}

// list is enveloped and byId is bare. Both shapes have to decode, and the
// per-kind store fields must land in the right places.
func TestListDocumentsHandlesBothKindsAndTheEnvelope(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		kind     DocumentKind
		body     string
		wantPath string
		checkDoc func(*testing.T, StoreDocument)
	}{
		{
			name: "writeoff list through the v2 envelope",
			kind: KindWriteoff,
			body: `{"result":"SUCCESS","revision":42,"response":[
			  {"id":"w-1","documentNumber":"W1","status":"PROCESSED","dateIncoming":"2026-03-02T10:00",
			   "storeId":"st-1","accountId":"ac-1",
			   "items":[{"num":1,"productId":"p-1","amount":2.5,"cost":100.5,"measureUnitId":"mu-1"}]}]}`,
			wantPath: EndpointWriteoffList,
			checkDoc: func(t *testing.T, d StoreDocument) {
				if d.StoreID != "st-1" || d.AccountID != "ac-1" {
					t.Errorf("writeoff store placement lost: %+v", d)
				}
				if d.Kind != KindWriteoff {
					t.Errorf("kind must be stamped by the client, got %q", d.Kind)
				}
				if len(d.Items) != 1 || d.Items[0].Amount != 2.5 || d.Items[0].Cost != 100.5 {
					t.Errorf("lines decoded wrong: %+v", d.Items)
				}
			},
		},
		{
			name: "internal transfer list as a bare array",
			kind: KindInternalTransfer,
			body: `[{"id":"t-1","documentNumber":"T1","status":"NEW","dateIncoming":"2026-03-03T11:30",
			  "storeFromId":"st-1","storeToId":"st-2","items":[{"productId":"p-9","amount":1}]}]`,
			wantPath: EndpointInternalTransferList,
			checkDoc: func(t *testing.T, d StoreDocument) {
				if d.StoreFromID != "st-1" || d.StoreToID != "st-2" {
					t.Errorf("transfer needs both stores: %+v", d)
				}
				if d.StoreID != "" {
					t.Errorf("a transfer has no single storeId, got %q", d.StoreID)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			var gotQuery url.Values
			s, c := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotQuery = r.URL.Path, r.URL.Query()
				_, _ = w.Write([]byte(tt.body))
			})
			defer c.Close(context.Background())

			from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
			to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
			got, err := s.ListDocuments(context.Background(), tt.kind, from, to, DocProcessed)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.wantPath) {
				t.Errorf("path %q, want suffix %q", gotPath, tt.wantPath)
			}
			if gotQuery.Get("dateFrom") != "2026-03-01" || gotQuery.Get("dateTo") != "2026-03-31" {
				t.Errorf("both date bounds are mandatory upstream, sent %v", gotQuery)
			}
			if gotQuery.Get("status") != DocProcessed {
				t.Errorf("status = %q, want %q", gotQuery.Get("status"), DocProcessed)
			}
			if len(got) != 1 {
				t.Fatalf("want one document, got %+v", got)
			}
			tt.checkDoc(t, got[0])
		})
	}
}

func TestListDocumentsOmitsABlankStatus(t *testing.T) {
	t.Parallel()
	var gotQuery url.Values
	s, c := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`[]`))
	})
	defer c.Close(context.Background())
	day := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if _, err := s.ListDocuments(context.Background(), KindWriteoff, day, day, "  "); err != nil {
		t.Fatalf("list: %v", err)
	}
	if _, present := gotQuery["status"]; present {
		t.Errorf("a blank status must be omitted, not sent empty: %v", gotQuery)
	}
}

func TestDocumentByIDDecodesTheBareObject(t *testing.T) {
	t.Parallel()
	var gotQuery url.Values
	s, c := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		_, _ = w.Write([]byte(`{"id":"w-7","status":"NEW","storeId":"st-3",
		  "items":[{"num":1,"productId":"p-1","amount":3}]}`))
	})
	defer c.Close(context.Background())
	got, err := s.DocumentByID(context.Background(), KindWriteoff, "w-7")
	if err != nil {
		t.Fatalf("byId: %v", err)
	}
	if got.ID != "w-7" || got.Kind != KindWriteoff || len(got.Items) != 1 {
		t.Errorf("decoded wrong: %+v", got)
	}
	if gotQuery.Get("id") != "w-7" {
		t.Errorf("id must go in the query, got %v", gotQuery)
	}
}

// A v2 business failure arrives as HTTP 200 with result=ERROR.
func TestListDocumentsSurfacesResultError(t *testing.T) {
	t.Parallel()
	s, c := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"result":"ERROR","errors":[{"code":"E","value":"period too long"}]}`))
	})
	defer c.Close(context.Background())
	day := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	_, err := s.ListDocuments(context.Background(), KindWriteoff, day, day, "")
	if err == nil || !strings.Contains(err.Error(), "period too long") {
		t.Errorf("result=ERROR on HTTP 200 must become an error, got %v", err)
	}
}
