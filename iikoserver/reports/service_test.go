package reports

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

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

func TestOlapRequestGuards(t *testing.T) {
	from, _ := rest.ParseDay("2026-03-01")
	to, _ := rest.ParseDay("2026-03-31")

	req, err := NewOlapRequest(ReportSales, from, to, []string{"OpenDate.Typed"}, nil, []string{"DishDiscountSumInt"})
	if err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	if req.BuildSummary {
		t.Fatal("buildSummary must default false — true can hang the server on large chains")
	}
	f, ok := req.Filters["OpenDate.Typed"]
	if !ok {
		t.Fatal("SALES must be filtered on OpenDate.Typed; iiko 5.5+ requires a date filter on every OLAP request")
	}
	if f.From != "2026-03-01T00:00:00.000" {
		t.Fatalf("OLAP filters need yyyy-MM-ddTHH:mm:ss.SSS, got %q", f.From)
	}
	if _, ok := req.Filters["DateTime.Typed"]; ok {
		t.Fatal("DateTime.Typed is the TRANSACTIONS date field, not SALES")
	}
	// Deleted positions must be excluded unless a caller opts in: revenue summed
	// from a report that includes them is overstated.
	for _, key := range []string{FilterDeletedWithWriteoff, FilterOrderDeleted} {
		f, ok := req.Filters[key]
		if !ok {
			t.Fatalf("SALES must exclude deleted positions by default; %s filter missing", key)
		}
		if f.FilterType != "IncludeValues" || len(f.Values) != 1 || f.Values[0] != "NOT_DELETED" {
			t.Fatalf("%s must be IncludeValues[NOT_DELETED], got %+v", key, f)
		}
	}

	tr, _ := NewOlapRequest(ReportTransactions, from, to, nil, nil, []string{"Sum.Incoming"})
	if _, ok := tr.Filters["DateTime.Typed"]; !ok {
		t.Fatal("TRANSACTIONS must be filtered on DateTime.Typed")
	}

	if _, err := NewOlapRequest(ReportSales, to, from, nil, nil, nil); err == nil {
		t.Fatal("reversed range must be rejected")
	}
	wide, _ := rest.ParseDay("2026-09-01")
	if _, err := NewOlapRequest(ReportSales, from, wide, nil, nil, nil); err == nil {
		t.Fatal("6-month range must be rejected: iiko documents a 1-month maximum")
	}
	many := []string{"a", "b", "c", "d"}
	if _, err := NewOlapRequest(ReportSales, from, to, many, many, many); err == nil {
		t.Fatal("12 fields must be rejected: iiko documents a maximum of 7")
	}
}

// A business failure arrives as HTTP 200 with result=ERROR in the body. Decoding
// that with a plain json.Unmarshal yields an empty report and looks like success,
// so a manager asking for sales would be told there were none. Every v2 JSON
// response, OLAP included, has to go through DecodeV2.
// AllowDeleted is the only way to see deleted rows, and it must not disturb the
// date filter the server requires on every request.
// catalogDTO mirrors the parts of research/dtos.yaml this test reads.
type catalogDTO struct {
	Name   string `yaml:"name"`
	Fields []struct {
		Name    string   `yaml:"name"`
		Aliases []string `yaml:"aliases"`
	} `yaml:"fields"`
}

// The alias lists in dateFieldGroups are a copy of what the catalog records, and
// a copy drifts. This pins them to research/dtos.yaml the way
// TestSalesEnumsExistInDocumentedCatalog pins the enum vocabularies: a deprecated
// spelling added to the catalog but not here would slip past applyFilters and
// produce a second date filter.
func TestDateFieldAliasesMatchCatalog(t *testing.T) {
	t.Parallel()
	path := catalogPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("catalog not present (%v) — alias drift unchecked", err)
	}
	var dtos []catalogDTO
	if err := yaml.Unmarshal(data, &dtos); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	// alias -> canonical, across every OLAP field DTO
	catalog := map[string]string{}
	for _, d := range dtos {
		if !strings.HasPrefix(d.Name, "OlapFields") {
			continue
		}
		for _, f := range d.Fields {
			for _, a := range f.Aliases {
				catalog[a] = f.Name
			}
		}
	}
	if len(catalog) == 0 {
		t.Fatal("no aliases found in the catalog — the loader is looking in the wrong place")
	}

	for _, rt := range []ReportType{ReportSales, ReportTransactions, ReportDeliveries} {
		t.Run(string(rt), func(t *testing.T) {
			t.Parallel()
			for canonical, aliases := range dateFieldGroups(rt) {
				for _, a := range aliases {
					got, ok := catalog[a]
					if !ok {
						t.Errorf("%s: %q is listed as an alias of %s but the catalog does not record it",
							rt, a, canonical)
						continue
					}
					if got != canonical {
						t.Errorf("%s: catalog says %q is an alias of %s, not %s", rt, a, got, canonical)
					}
				}
			}
			// Every catalogued alias of a field we treat as this report's date
			// field must be covered, or it slips past the filter check.
			for alias, canonical := range catalog {
				if _, isDate := dateFieldGroups(rt)[canonical]; !isDate {
					continue
				}
				if _, covered := rt.CanonicalDateField(alias); !covered {
					t.Errorf("%s: catalog records %q as an alias of the date field %s, but it is not blocked",
						rt, alias, canonical)
				}
			}
		})
	}
}

func TestCanonicalDateFieldMatchesExactlyNotByPrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		rt        ReportType
		field     string
		wantCanon string
		wantDate  bool
	}{
		{name: "SALES canonical", rt: ReportSales, field: "OpenDate.Typed", wantCanon: "OpenDate.Typed", wantDate: true},
		{name: "SALES deprecated OpenDate", rt: ReportSales, field: "OpenDate", wantCanon: "OpenDate.Typed", wantDate: true},
		{name: "SALES deprecated OperDayFilter", rt: ReportSales, field: "OpenDate.OperDayFilter", wantCanon: "OpenDate.Typed", wantDate: true},
		{name: "DELIVERIES shares the SALES field", rt: ReportDeliveries, field: "OpenDate", wantCanon: "OpenDate.Typed", wantDate: true},
		{name: "TRANSACTIONS canonical", rt: ReportTransactions, field: "DateTime.Typed", wantCanon: "DateTime.Typed", wantDate: true},
		{name: "TRANSACTIONS deprecated DateTime", rt: ReportTransactions, field: "DateTime", wantCanon: "DateTime.Typed", wantDate: true},
		{name: "TRANSACTIONS date-only variant", rt: ReportTransactions, field: "DateTime.DateTyped", wantCanon: "DateTime.DateTyped", wantDate: true},
		{name: "TRANSACTIONS deprecated DateTime.Date", rt: ReportTransactions, field: "DateTime.Date", wantCanon: "DateTime.DateTyped", wantDate: true},

		// Exact matching: a field that merely starts the same way is ordinary.
		{name: "prefix lookalike is not a date field", rt: ReportSales, field: "OpenDateSomethingElse"},
		{name: "delivery customer date is a real filterable column", rt: ReportDeliveries, field: "Delivery.CustomerCreatedDateTyped"},
		{name: "the other type's field is not this type's", rt: ReportSales, field: "DateTime.Typed"},
		{name: "ordinary column", rt: ReportSales, field: "WaiterName"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			canon, isDate := tt.rt.CanonicalDateField(tt.field)
			if isDate != tt.wantDate {
				t.Fatalf("CanonicalDateField(%q) date = %v, want %v", tt.field, isDate, tt.wantDate)
			}
			if canon != tt.wantCanon {
				t.Errorf("CanonicalDateField(%q) canonical = %q, want %q", tt.field, canon, tt.wantCanon)
			}
		})
	}
}

// DateField must always be among the spellings the filter check blocks.
func TestDateFieldIsCoveredByItsOwnAliasList(t *testing.T) {
	t.Parallel()
	for _, rt := range []ReportType{ReportSales, ReportTransactions, ReportDeliveries} {
		aliases := rt.DateFieldAliases()
		found := false
		for _, a := range aliases {
			if a == rt.DateField() {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: DateField %q missing from DateFieldAliases %v", rt, rt.DateField(), aliases)
		}
	}
}

func TestAllowDeletedDropsOnlyTheDeletionFilters(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	to, _ := rest.ParseDay("2026-03-31")

	tests := []struct {
		name        string
		rt          ReportType
		allow       bool
		wantDeleted bool
	}{
		{name: "SALES excludes deleted by default", rt: ReportSales, wantDeleted: true},
		{name: "SALES with AllowDeleted keeps no deletion filter", rt: ReportSales, allow: true},
		{name: "TRANSACTIONS never had them", rt: ReportTransactions},
		{name: "TRANSACTIONS AllowDeleted is a no-op", rt: ReportTransactions, allow: true},
		{name: "DELIVERIES never had them", rt: ReportDeliveries},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req, err := NewOlapRequest(tt.rt, from, to, nil, nil, []string{"DishSumInt"})
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			if tt.allow {
				req.AllowDeleted()
			}
			_, gotWriteoff := req.Filters[FilterDeletedWithWriteoff]
			_, gotOrder := req.Filters[FilterOrderDeleted]
			if gotWriteoff != tt.wantDeleted || gotOrder != tt.wantDeleted {
				t.Errorf("deletion filters present = (%v,%v), want %v", gotWriteoff, gotOrder, tt.wantDeleted)
			}
			// the date filter is mandatory on every request, deleted or not
			if _, ok := req.Filters[tt.rt.DateField()]; !ok {
				t.Errorf("AllowDeleted must not disturb the %s date filter", tt.rt.DateField())
			}
		})
	}
}

func TestOlapSurfacesResultErrorOnHTTP200(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		body    string
		wantErr bool
		wantSub string
	}{
		{
			name:    "result ERROR with messages becomes an error",
			body:    `{"result":"ERROR","errors":[{"code":"E1","value":"period too long"}]}`,
			wantErr: true,
			wantSub: "period too long",
		},
		{
			name:    "result ERROR with plain string errors becomes an error",
			body:    `{"result":"ERROR","errors":["no licence for module"]}`,
			wantErr: true,
			wantSub: "no licence",
		},
		{
			name: "bare unenveloped report still decodes",
			body: `{"data":[{"DishName":"Tea"}],"summary":[]}`,
		},
		{
			name: "SUCCESS envelope unwraps to the report",
			body: `{"result":"SUCCESS","response":{"data":[{"DishName":"Tea"}],"summary":[]}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, c := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.body))
			})
			defer c.Close(context.Background())
			from, _ := rest.ParseDay("2026-01-01")
			to, _ := rest.ParseDay("2026-01-02")
			req, err := NewOlapRequest(ReportSales, from, to, []string{"DishName"}, nil, []string{"DishSumInt"})
			if err != nil {
				t.Fatalf("build request: %v", err)
			}
			got, err := s.Olap(context.Background(), req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("want an error for %s, got report %+v", tt.body, got)
				}
				if !strings.Contains(err.Error(), tt.wantSub) {
					t.Errorf("error %q must carry iiko's own message %q", err, tt.wantSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.Data) != 1 {
				t.Errorf("want 1 row decoded, got %d", len(got.Data))
			}
		})
	}
}

func TestOlapFieldsFiltersToTheRequestedReport(t *testing.T) {
	t.Parallel()
	var gotQ url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		_, _ = w.Write([]byte(`{"DishName":{"name":"Блюдо","type":"STRING","groupingAllowed":true,"filteringAllowed":true},
		                        "DishSumInt":{"name":"Сумма","type":"MONEY","aggregationAllowed":true}}`))
	})
	got, err := s.OlapFields(context.Background(), ReportTransactions)
	if err != nil {
		t.Fatal(err)
	}
	if gotQ.Get("reportType") != "TRANSACTIONS" {
		t.Errorf("reportType = %q, want TRANSACTIONS", gotQ.Get("reportType"))
	}
	if len(got) != 2 || !got["DishName"].GroupingAllow || !got["DishSumInt"].AggregationAllow {
		t.Fatalf("capability flags lost in decode: %+v", got)
	}
}

func TestStoreBalancesUsesTimestampFormatAndRepeatsFilters(t *testing.T) {
	t.Parallel()
	at, _ := rest.ParseDay("2026-03-05")
	var gotQ url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		_, _ = w.Write([]byte(`[{"store":"Кухня","product":"Мука","amount":12.5,"sum":900}]`))
	})
	bal, err := s.StoreBalances(context.Background(), at, []string{"s1", "s2"}, []string{"p1"})
	if err != nil {
		t.Fatal(err)
	}
	// This endpoint takes yyyy-MM-dd'T'HH:mm:ss, not the plain date other v2 params use.
	if got := gotQ.Get("timestamp"); got != "2026-03-05T00:00:00" {
		t.Errorf("timestamp = %q, want 2026-03-05T00:00:00", got)
	}
	// Multi-value params repeat the key rather than joining with commas.
	if got := gotQ["store"]; len(got) != 2 || got[0] != "s1" || got[1] != "s2" {
		t.Errorf("store filters = %v, want two repeated keys", got)
	}
	if len(bal) != 1 || bal[0].Amount != 12.5 {
		t.Fatalf("bad decode: %+v", bal)
	}
}

// catalogPath walks up to the module root rather than counting "..": this test
// skipped silently for a whole refactor when the package moved deeper.
func catalogPath(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "research", "dtos.yaml")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod above the test's working directory")
		}
		dir = parent
	}
}
