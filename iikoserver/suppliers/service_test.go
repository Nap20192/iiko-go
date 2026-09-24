package suppliers

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

func newFake(t *testing.T, h http.HandlerFunc) (*Service, *rest.Client) {
	t.Helper()
	c := resttest.NewFake(t, h)
	return New(c), c
}

const suppliersXML = `<employees>
 <employee><id>s-1</id><code>SUP1</code><name>ООО Ромашка</name><deleted>false</deleted></employee>
 <employee><id>s-2</id><code>SUP2</code><name>Старый</name><deleted>true</deleted></employee>
</employees>`

// Suppliers are Users with supplier=true, so this endpoint answers the employee
// XML shape rather than a supplier-specific one.
func TestListDecodesTheEmployeeShapeAndFiltersDeleted(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		incl bool
		want int
	}{
		{name: "deleted excluded", incl: false, want: 1},
		{name: "deleted included", incl: true, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotQ url.Values
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotQ, gotPath = r.URL.Query(), r.URL.Path
				_, _ = w.Write([]byte(suppliersXML))
			})
			got, err := s.List(context.Background(), tt.incl)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("got %d suppliers, want %d", len(got), tt.want)
			}
			if !strings.HasSuffix(gotPath, EndpointSuppliers) {
				t.Errorf("path = %q", gotPath)
			}
			if gotQ.Get("includeDeleted") != rest.BoolStr(tt.incl) {
				t.Errorf("includeDeleted must be explicit, got %q", gotQ.Get("includeDeleted"))
			}
		})
	}
}

// Search has no id field — the docs list every other identifier but not that one,
// so asking by id silently matches nothing.
func TestSearchRejectsAnIDField(t *testing.T) {
	t.Parallel()
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not spend a request on a field iiko does not search")
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := s.Search(context.Background(), map[string]string{"id": "s-1"}); err == nil ||
		!strings.Contains(err.Error(), "id") {
		t.Errorf("searching by id is unsupported and must be refused, got %v", err)
	}
	if _, err := s.Search(context.Background(), nil); err == nil {
		t.Error("an unfiltered supplier search must be refused")
	}
}

// The pricelist takes a supplier CODE in the path, not a UUID, and its date is
// DD.MM.YYYY while its neighbours are ISO.
func TestPricelistUsesACodeAndTheDottedDate(t *testing.T) {
	t.Parallel()
	on, _ := rest.ParseDay("2026-03-05")
	tests := []struct {
		name     string
		onDate   bool
		wantDate string
	}{
		{name: "latest when no date", wantDate: ""},
		{name: "as of a date", onDate: true, wantDate: "05.03.2026"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			var gotQ url.Values
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotQ = r.URL.Path, r.URL.Query()
				_, _ = w.Write([]byte(`<priceListItems><item><productId>p-1</productId><price>10.5</price></item></priceListItems>`))
			})
			var when *time.Time
			if tt.onDate {
				when = &on
			}
			got, err := s.Pricelist(context.Background(), "SUP1", when)
			if err != nil {
				t.Fatalf("pricelist: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("decoded %d items", len(got))
			}
			if !strings.HasSuffix(gotPath, "/api/suppliers/SUP1/pricelist") {
				t.Errorf("the supplier CODE goes in the path, got %q", gotPath)
			}
			if gotQ.Get("date") != tt.wantDate {
				t.Errorf("date = %q, want %q (DD.MM.YYYY here, ISO next door)", gotQ.Get("date"), tt.wantDate)
			}
		})
	}
}

func TestPricelistRequiresACode(t *testing.T) {
	t.Parallel()
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not reach iiko without a supplier code")
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := s.Pricelist(context.Background(), " ", nil); err == nil {
		t.Error("a blank supplier code must be refused")
	}
}
