package edi

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
	"github.com/Nap20192/iiko-go/iikoserver/rest/resttest"
)

const system = "edi-guid-1"

type call struct {
	method string
	url    url.URL
	body   string
}

func newFake(t *testing.T, response string) (*Service, *call) {
	t.Helper()
	got := &call{}
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*got = call{method: r.Method, url: *r.URL, body: string(b)}
		_, _ = w.Write([]byte(response))
	})
	return New(c), got
}

func day(s string) time.Time {
	t, err := time.Parse(rest.QueryV2, s)
	if err != nil {
		panic(err)
	}
	return t
}

func ids() []DocumentIdentifierDto {
	return []DocumentIdentifierDto{{Number: "10001", Date: "2026-06-02"}}
}

// TestVerbsMatchTheCatalog is the point of this domain's test file. The docs'
// rendered request blocks show POST for most of these; the CSS classes they are
// derived from say PUT, and research/verbs-from-css.tsv records which. A wrong
// verb here does not fail loudly — iiko answers 404 or silently does nothing.
func TestVerbsMatchTheCatalog(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		want   string
		path   string
		reply  string
		invoke func(*Service) error
	}{
		{"invoice", http.MethodPut, EndpointInvoice, ``, func(s *Service) error {
			return s.SendInvoice(context.Background(), system, EdiMessageDto{ID: "m-1"})
		}},
		{"orders/ack", http.MethodPut, EndpointOrdersAck, ``, func(s *Service) error {
			return s.AckOrder(context.Background(), system, "10001", day("2026-06-02"), "")
		}},
		{"orders/bySeller", http.MethodGet, EndpointOrdersBySeller, `<ediMessageDtoes/>`, func(s *Service) error {
			_, err := s.OrdersBySeller(context.Background(), system, SellerFilter{INN: "7701234567"})
			return err
		}},
		{"orders/create", http.MethodPost, EndpointOrdersCreate, `<orderCreationOrUpdateResultDto/>`, func(s *Service) error {
			_, err := s.CreateOrder(context.Background(), system, OrderDto{Date: "2026-06-02"})
			return err
		}},
		{"orders/list", http.MethodGet, EndpointOrdersList, `<orderDtoes/>`, func(s *Service) error {
			_, err := s.ListOrders(context.Background(), system, time.Time{}, time.Time{}, -1)
			return err
		}},
		{"orders/register", http.MethodPut, EndpointOrdersRegister, `<orderDtoes/>`, func(s *Service) error {
			_, err := s.RegisterOrders(context.Background(), system, ids())
			return err
		}},
		{"orders/send", http.MethodPut, EndpointOrdersSend, `<orderDtoes/>`, func(s *Service) error {
			_, err := s.SendOrders(context.Background(), system, ids())
			return err
		}},
		{"orders/unregister", http.MethodPut, EndpointOrdersUnregister, `<orderDtoes/>`, func(s *Service) error {
			_, err := s.UnregisterOrders(context.Background(), system, ids())
			return err
		}},
		{"orders/update", http.MethodPut, EndpointOrdersUpdate, `<orderCreationOrUpdateResultDto/>`, func(s *Service) error {
			_, err := s.UpdateOrder(context.Background(), system, OrderDto{Number: "10001", Date: "2026-06-02"})
			return err
		}},
		{"response", http.MethodPut, EndpointResponse, `<ediMessage/>`, func(s *Service) error {
			_, err := s.RespondToOrder(context.Background(), system, EdiMessageDto{ID: "m-1"})
			return err
		}},
	}
	if len(cases) != 10 {
		t.Fatalf("the catalog lists 10 EDI endpoints, this table has %d", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, got := newFake(t, tc.reply)
			if err := tc.invoke(s); err != nil {
				t.Fatal(err)
			}
			if got.method != tc.want {
				t.Errorf("used %s, catalog says %s", got.method, tc.want)
			}
			want := "/resto" + strings.Replace(tc.path, "%s", system, 1)
			if got.url.Path != want {
				t.Errorf("path %s, want %s", got.url.Path, want)
			}
		})
	}
}

func TestAckRequiresTheIdentityKeyAndDefaultsToOriginal(t *testing.T) {
	t.Parallel()

	s, got := newFake(t, ``)
	if err := s.AckOrder(context.Background(), system, "10001", day("2026-06-02"), ""); err != nil {
		t.Fatal(err)
	}
	q := got.url.Query()
	if q.Get("number") != "10001" || q.Get("date") != "2026-06-02" {
		t.Errorf("query was %s", got.url.RawQuery)
	}
	// An unacknowledged order keeps reappearing in bySeller, so the default has
	// to be the one that acknowledges receipt.
	if q.Get("status") != AckOriginal {
		t.Errorf("status = %q, want %q", q.Get("status"), AckOriginal)
	}

	if err := s.AckOrder(context.Background(), system, "", day("2026-06-02"), ""); err == nil {
		t.Error("number is half the identity key and is required")
	}
	if err := s.AckOrder(context.Background(), system, "10001", time.Time{}, ""); err == nil {
		t.Error("date is the other half and is required")
	}
	if err := s.AckOrder(context.Background(), system, "10001", day("2026-06-02"), "acked"); err == nil {
		t.Error("only original and canceled are documented")
	}
}

func TestBySellerNeedsGLNOrINN(t *testing.T) {
	t.Parallel()

	s, got := newFake(t, `<ediMessageDtoes><ediMessage id="m-1">
		<header><documentType>ORDERS</documentType></header></ediMessage></ediMessageDtoes>`)
	msgs, err := s.OrdersBySeller(context.Background(), system, SellerFilter{GLN: "4600000000000", KPP: "770101001"})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].Header.DocumentType != "ORDERS" {
		t.Fatalf("got %+v", msgs)
	}
	if got.url.Query().Get("gln") != "4600000000000" || got.url.Query().Get("kpp") != "770101001" {
		t.Errorf("query was %s", got.url.RawQuery)
	}
	// Without a seller the call returns every participant's orders, which the
	// docs treat as a filter error rather than a wildcard.
	if _, err := s.OrdersBySeller(context.Background(), system, SellerFilter{Name: "ООО Поставщик"}); err == nil {
		t.Error("gln or inn is required")
	}
}

func TestListOrdersSendsTheRevisionCursorAndOptionalWindow(t *testing.T) {
	t.Parallel()

	s, got := newFake(t, `<orderDtoes><order number="10001" processingStatus="registered"/></orderDtoes>`)
	orders, err := s.ListOrders(context.Background(), system, day("2026-06-01"), day("2026-06-30"), 42)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 1 || orders[0].ProcessingStatus != "registered" {
		t.Fatalf("got %+v", orders)
	}
	q := got.url.Query()
	if q.Get("from") != "2026-06-01" || q.Get("to") != "2026-06-30" || q.Get("revisionFrom") != "42" {
		t.Errorf("query was %s", got.url.RawQuery)
	}

	// -1 is the documented "give me everything" cursor, not "omit the parameter".
	if _, err := s.ListOrders(context.Background(), system, time.Time{}, time.Time{}, -1); err != nil {
		t.Fatal(err)
	}
	q = got.url.Query()
	if q.Get("revisionFrom") != "-1" {
		t.Errorf("revisionFrom = %q", q.Get("revisionFrom"))
	}
	if _, ok := q["from"]; ok {
		t.Error("an absent window must not be sent as a zero date")
	}
}

func TestRegisterSendsTheDocumentedIdentifierArray(t *testing.T) {
	t.Parallel()

	s, got := newFake(t, `<orderDtoes><order number="10001" processingStatus="registered"/></orderDtoes>`)
	orders, err := s.RegisterOrders(context.Background(), system, ids())
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 1 {
		t.Fatalf("got %+v", orders)
	}
	for _, want := range []string{
		"<documentIdentifierDtoes>",
		`<documentIdentifier number="10001" date="2026-06-02">`,
	} {
		if !strings.Contains(got.body, want) {
			t.Errorf("missing %s in %s", want, got.body)
		}
	}
	if _, err := s.RegisterOrders(context.Background(), system, nil); err == nil {
		t.Error("an empty identifier list would register nothing and read as success")
	}
}

func TestUpdateRefusesToDropTheImmutableKey(t *testing.T) {
	t.Parallel()

	// "Номер и дата документа не могут быть модифицированы, т.к. используются
	// для идентификации заказа." Sending an order without them does not update
	// anything — it addresses no document.
	s, _ := newFake(t, `<orderCreationOrUpdateResultDto/>`)
	if _, err := s.UpdateOrder(context.Background(), system, OrderDto{Date: "2026-06-02"}); err == nil {
		t.Error("number is required on update")
	}
	if _, err := s.UpdateOrder(context.Background(), system, OrderDto{Number: "10001"}); err == nil {
		t.Error("date is required on update")
	}
}

func TestCreateStripsTheServerOwnedProcessingStatus(t *testing.T) {
	t.Parallel()

	s, got := newFake(t, `<orderCreationOrUpdateResultDto>
		<warnings><warning code="SUPPLIER_PRODUCT_NOT_FOUND">нет в прайсе</warning></warnings>
		<order number="10007" processingStatus="draft"/>
		</orderCreationOrUpdateResultDto>`)
	res, err := s.CreateOrder(context.Background(), system, OrderDto{
		Date: "2026-06-02", ProcessingStatus: "registered",
	})
	if err != nil {
		t.Fatal(err)
	}
	// A created order always starts as draft; asking for anything else is not
	// refused, it is ignored, which is worse.
	if strings.Contains(got.body, "registered") {
		t.Errorf("processingStatus was sent: %s", got.body)
	}
	if len(res.Warnings) != 1 || res.Warnings[0].Code != "SUPPLIER_PRODUCT_NOT_FOUND" {
		t.Fatalf("warnings are the whole point of the response: %+v", res)
	}
	if res.Order.Number != "10007" {
		t.Errorf("server-assigned number lost: %+v", res.Order)
	}
}

func TestCancelLineZeroesBothQuantities(t *testing.T) {
	t.Parallel()

	// Cancelling a confirmed line is "resend it with confirmedQuantity zero".
	// Leaving containerQuantity set cancels the units but still ships the boxes.
	line := OrderResponseItemDto{
		LineNumber:        3,
		OrderLineNumber:   3,
		ConfirmedQuantity: QuantityDto{MeasureUnit: "KG", Quantity: 12, ContainerQuantity: 2},
	}
	got := CancelLine(line)
	if got.ConfirmedQuantity.Quantity != 0 || got.ConfirmedQuantity.ContainerQuantity != 0 {
		t.Fatalf("got %+v", got.ConfirmedQuantity)
	}
	if got.ConfirmedQuantity.MeasureUnit != "KG" || got.LineNumber != 3 {
		t.Fatalf("the rest of the line must be resent unchanged: %+v", got)
	}
	// The input must not be mutated: a caller usually cancels one line of many.
	if line.ConfirmedQuantity.Quantity != 12 {
		t.Error("CancelLine mutated its argument")
	}
}

func TestEveryCallRejectsAnEmptyEdiSystem(t *testing.T) {
	t.Parallel()

	// The GUID sits in the middle of every path. Empty, it produces
	// /api/edi//orders/list, which iiko answers with a 404 that reads like a
	// missing endpoint rather than a missing configuration value.
	s, _ := newFake(t, `<orderDtoes/>`)
	calls := map[string]func() error{
		"SendInvoice":      func() error { return s.SendInvoice(context.Background(), "", EdiMessageDto{}) },
		"AckOrder":         func() error { return s.AckOrder(context.Background(), "", "1", day("2026-06-02"), "") },
		"OrdersBySeller":   func() error { _, e := s.OrdersBySeller(context.Background(), "", SellerFilter{INN: "1"}); return e },
		"CreateOrder":      func() error { _, e := s.CreateOrder(context.Background(), "", OrderDto{}); return e },
		"ListOrders":       func() error { _, e := s.ListOrders(context.Background(), "", time.Time{}, time.Time{}, -1); return e },
		"RegisterOrders":   func() error { _, e := s.RegisterOrders(context.Background(), "", ids()); return e },
		"SendOrders":       func() error { _, e := s.SendOrders(context.Background(), "", ids()); return e },
		"UnregisterOrders": func() error { _, e := s.UnregisterOrders(context.Background(), "", ids()); return e },
		"UpdateOrder": func() error {
			_, e := s.UpdateOrder(context.Background(), "", OrderDto{Number: "1", Date: "d"})
			return e
		},
		"RespondToOrder": func() error { _, e := s.RespondToOrder(context.Background(), "", EdiMessageDto{}); return e },
	}
	if len(calls) != 10 {
		t.Fatalf("the catalog lists 10 EDI endpoints, this table has %d", len(calls))
	}
	for name, call := range calls {
		if err := call(); err == nil || !strings.Contains(err.Error(), "ediSystem") {
			t.Errorf("%s: got %v, want an ediSystem error", name, err)
		}
	}
}

// TestAttributeFormRoundTrips pins what the catalog now generates against the
// docs' own worked examples: identity fields are attributes and list fields wrap
// their children. Getting this wrong fails silently in both directions — iiko
// ignores unknown elements on a write, and a read decodes an empty order.
func TestAttributeFormRoundTrips(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		xml  string
		into func() any
		want func(any) bool
	}{
		{
			name: "order",
			xml: `<order number="10004" date="2016-08-15" status="original" processingStatus="readyToSend">` +
				`<comment>срочно</comment>` +
				`<lineItems><lineItem><lineNumber>1</lineNumber><name>Апельсин</name></lineItem></lineItems>` +
				`</order>`,
			into: func() any { return new(OrderDto) },
			want: func(v any) bool {
				o := v.(*OrderDto)
				return o.Number == "10004" && o.Date == "2016-08-15" && o.Status == "original" &&
					o.ProcessingStatus == "readyToSend" && o.Comment == "срочно" &&
					len(o.LineItems) == 1 && o.LineItems[0].Name == "Апельсин"
			},
		},
		{
			name: "documentIdentifier",
			xml:  `<documentIdentifier number="10026" date="2016-08-24"></documentIdentifier>`,
			into: func() any { return new(DocumentIdentifierDto) },
			want: func(v any) bool {
				d := v.(*DocumentIdentifierDto)
				return d.Number == "10026" && d.Date == "2016-08-24"
			},
		},
		{
			name: "ediMessage",
			xml: `<ediMessage id="3eacd707-ad57-9161-0155-1178b8a00010" creationDateTime="2016-06-02T17:17:05">` +
				`<header><documentType>ORDERS</documentType></header></ediMessage>`,
			into: func() any { return new(EdiMessageDto) },
			want: func(v any) bool {
				m := v.(*EdiMessageDto)
				return m.ID == "3eacd707-ad57-9161-0155-1178b8a00010" &&
					m.CreationDateTime == "2016-06-02T17:17:05" && m.Header.DocumentType == "ORDERS"
			},
		},
		{
			name: "warning",
			xml:  `<warning code="SUPPLIER_PRODUCT_NOT_FOUND">Supplier product with code 1515151511 not found</warning>`,
			into: func() any { return new(EdiOrderWarningDto) },
			want: func(v any) bool {
				w := v.(*EdiOrderWarningDto)
				return w.Code == "SUPPLIER_PRODUCT_NOT_FOUND" && strings.Contains(w.Message, "1515151511")
			},
		},
		{
			name: "ordersp",
			xml: `<ordersp number="10001" date="2016-06-02" status="Accepted">` +
				`<lineItems><lineItem><lineNumber>1</lineNumber>` +
				`<confirmedQuantity><measureUnit>KG</measureUnit><quantity>2.5</quantity></confirmedQuantity>` +
				`</lineItem></lineItems></ordersp>`,
			into: func() any { return new(OrderResponseDto) },
			want: func(v any) bool {
				r := v.(*OrderResponseDto)
				return r.Number == "10001" && r.Status == "Accepted" &&
					len(r.LineItems) == 1 && r.LineItems[0].ConfirmedQuantity.Quantity == 2.5
			},
		},
		{
			name: "orderCreationOrUpdateResult",
			xml: `<orderCreationOrUpdateResultDto>` +
				`<warnings><warning code="A">one</warning></warnings>` +
				`<order number="10026" date="2016-08-24" processingStatus="draft"></order>` +
				`</orderCreationOrUpdateResultDto>`,
			into: func() any { return new(OrderCreationOrUpdateResultDto) },
			want: func(v any) bool {
				r := v.(*OrderCreationOrUpdateResultDto)
				return len(r.Warnings) == 1 && r.Order.Number == "10026" && r.Order.ProcessingStatus == "draft"
			},
		},
		{
			name: "invoice",
			xml: `<invoice number="INV-1" date="2016-06-03" type="INVOIC">` +
				`<originOrder number="10001" date="2016-06-02"></originOrder>` +
				`<lineItems><lineItem><lineNumber>1</lineNumber></lineItem></lineItems></invoice>`,
			into: func() any { return new(InvoiceDto) },
			want: func(v any) bool {
				i := v.(*InvoiceDto)
				return i.Number == "INV-1" && i.Type == "INVOIC" &&
					i.OriginOrder.Number == "10001" && len(i.LineItems) == 1
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			decoded := tc.into()
			if err := xml.Unmarshal([]byte(tc.xml), decoded); err != nil {
				t.Fatal(err)
			}
			if !tc.want(decoded) {
				t.Fatalf("decoded wrong: %+v", decoded)
			}

			// Re-encoding must produce attributes again, or a write silently
			// sends a document the server reads as empty.
			out, err := xml.Marshal(decoded)
			if err != nil {
				t.Fatal(err)
			}
			again := tc.into()
			if err := xml.Unmarshal(out, again); err != nil {
				t.Fatalf("re-encoded form does not parse: %s", out)
			}
			if !reflect.DeepEqual(decoded, again) {
				t.Fatalf("round trip lost data:\n in %+v\nout %+v\nxml %s", decoded, again, out)
			}
		})
	}
}

func TestWarningsWrapTheirChildren(t *testing.T) {
	t.Parallel()

	// <warnings> is a wrapper, not a repeated element; reading it flat drops
	// every warning, and warnings are the only signal that a created order
	// references products the supplier does not sell.
	const body = `<orderCreationOrUpdateResultDto><warnings>` +
		`<warning code="A">one</warning><warning code="B">two</warning>` +
		`</warnings><order number="1"/></orderCreationOrUpdateResultDto>`

	var res OrderCreationOrUpdateResultDto
	if err := xml.Unmarshal([]byte(body), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Warnings) != 2 || res.Warnings[1].Code != "B" || res.Warnings[1].Message != "two" {
		t.Fatalf("got %+v", res.Warnings)
	}
}
