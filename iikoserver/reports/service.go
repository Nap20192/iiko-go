package reports

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the OLAP and balance reports half of the iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// ReportType is the v2 OLAP vocabulary. v2 has exactly three — STOCK exists
// only in OLAP v1 and as a preset type, never here.
type ReportType string

const (
	ReportSales        ReportType = "SALES"
	ReportTransactions ReportType = "TRANSACTIONS"
	ReportDeliveries   ReportType = "DELIVERIES"
)

// DateField returns the field this report type must be filtered on. Since iiko
// 5.5 every OLAP request must carry a date filter, so this is not optional.
func (r ReportType) DateField() string {
	if r == ReportTransactions {
		return "DateTime.Typed"
	}
	return "OpenDate.Typed"
}

// dateFieldGroups lists, per report type, each field that acts as the report's
// period filter together with its deprecated spellings. Old presets still carry
// the deprecated names and the server still accepts them, so a caller filtering
// on one would add a second date filter fighting the period.
//
// The alias lists mirror research/dtos.yaml and are checked against it by
// TestDateFieldAliasesMatchCatalog.
//
// TRANSACTIONS has two such fields: DateTime.Typed (the one DateField returns)
// and DateTime.DateTyped, the date-only variant. Both filter the period.
// DELIVERIES shares the SALES field.
func dateFieldGroups(r ReportType) map[string][]string {
	if r == ReportTransactions {
		return map[string][]string{
			"DateTime.Typed":     {"DateTime"},
			"DateTime.DateTyped": {"DateTime.Date", "DateTime.OperDayFilter"},
		}
	}
	return map[string][]string{
		"OpenDate.Typed": {"OpenDate", "OpenDate.OperDayFilter"},
	}
}

// DateFieldAliases returns every documented spelling of this report type's date
// field or fields: the canonical names plus their deprecated aliases.
func (r ReportType) DateFieldAliases() []string {
	groups := dateFieldGroups(r)
	out := make([]string, 0, len(groups)*3)
	for canonical, aliases := range groups {
		out = append(out, canonical)
		out = append(out, aliases...)
	}
	sort.Strings(out) // stable order: callers render this in error messages
	return out
}

// CanonicalDateField reports whether field is one of this report type's date
// fields under any documented spelling, and returns the canonical name it maps
// to. Matching is exact: OpenDate.Typed is a date field, OpenDateSomethingElse
// is an ordinary column and must stay filterable.
func (r ReportType) CanonicalDateField(field string) (string, bool) {
	for canonical, aliases := range dateFieldGroups(r) {
		if field == canonical {
			return canonical, true
		}
		for _, a := range aliases {
			if field == a {
				return canonical, true
			}
		}
	}
	return "", false
}

type Filter struct {
	FilterType  string   `json:"filterType"`           // DateRange | IncludeValues | ExcludeValues | Range
	PeriodType  string   `json:"periodType,omitempty"` // CUSTOM, TODAY, LAST_WEEK, ...
	From        string   `json:"from,omitempty"`       // required even when periodType != CUSTOM
	To          string   `json:"to,omitempty"`
	IncludeLow  *bool    `json:"includeLow,omitempty"`  // defaults true
	IncludeHigh *bool    `json:"includeHigh,omitempty"` // defaults false
	Values      []string `json:"values,omitempty"`
}

type OlapRequest struct {
	ReportType       ReportType        `json:"reportType"`
	BuildSummary     bool              `json:"buildSummary"`
	GroupByRowFields []string          `json:"groupByRowFields"`
	GroupByColFields []string          `json:"groupByColFields"`
	AggregateFields  []string          `json:"aggregateFields"`
	Filters          map[string]Filter `json:"filters"`
}

type OlapResponse struct {
	Data    []map[string]any `json:"data"`
	Summary []any            `json:"summary"`
}

// MaxOlapFields is the documented ceiling: "используйте не более 7 полей".
// It counts grouping and aggregate fields; filters are not fields in that sense.
const MaxOlapFields = 7

// The two filters that keep deleted orders out of a SALES report. Named so the
// default and the opt-out cannot drift apart.
const (
	FilterDeletedWithWriteoff = "DeletedWithWriteoff"
	FilterOrderDeleted        = "OrderDeleted"
)

// NewOlapRequest builds a well-formed request with the date filter the server
// requires and the two deletion filters every sane sales report wants.
// buildSummary defaults false deliberately: true "может привести к зависанию
// сервера" on large chains.
func NewOlapRequest(rt ReportType, from, to time.Time, rows, cols, aggs []string) (*OlapRequest, error) {
	// v2 has exactly three report types. STOCK is v1-only, and sending it here
	// is answered with an unhelpful 400 rather than a named error.
	switch rt {
	case ReportSales, ReportTransactions, ReportDeliveries:
	default:
		return nil, fmt.Errorf("OLAP v2 has no %q report; valid: SALES, TRANSACTIONS, DELIVERIES (STOCK exists only on the v1 report)", rt)
	}
	if n := len(rows) + len(cols) + len(aggs); n > MaxOlapFields {
		return nil, fmt.Errorf("OLAP request uses %d fields; iiko documents a maximum of %d. Drop some group_by or metrics", n, MaxOlapFields)
	}
	if !to.After(from) {
		return nil, fmt.Errorf("date_to (%s) must be after date_from (%s)", to.Format(rest.QueryV2), from.Format(rest.QueryV2))
	}
	if to.Sub(from) > 31*24*time.Hour {
		return nil, fmt.Errorf("range is %d days; iiko recommends at most one month (ideally a day or a week). Narrow date_from/date_to", int(to.Sub(from).Hours()/24))
	}
	yes, no := true, false
	req := &OlapRequest{
		ReportType:       rt,
		BuildSummary:     false,
		GroupByRowFields: rows,
		GroupByColFields: cols,
		AggregateFields:  aggs,
		Filters: map[string]Filter{
			rt.DateField(): {
				FilterType: "DateRange", PeriodType: "CUSTOM",
				From: from.Format(rest.OlapMs), To: to.Format(rest.OlapMs),
				IncludeLow: &yes, IncludeHigh: &no,
			},
		},
	}
	if rt == ReportSales {
		req.Filters[FilterDeletedWithWriteoff] = Filter{FilterType: "IncludeValues", Values: []string{"NOT_DELETED"}}
		req.Filters[FilterOrderDeleted] = Filter{FilterType: "IncludeValues", Values: []string{"NOT_DELETED"}}
	}
	return req, nil
}

// AllowDeleted drops the deletion filters NewOlapRequest applies to a SALES
// report, so deleted and written-off positions appear in the result.
//
// It is a separate step rather than a constructor argument so that excluding
// them stays the default: a caller has to say the word to see deleted rows, and
// revenue summed from a request that includes them is overstated.
//
// A no-op for report types that never carried the filters.
func (r *OlapRequest) AllowDeleted() {
	delete(r.Filters, FilterDeletedWithWriteoff)
	delete(r.Filters, FilterOrderDeleted)
}

func (s *Service) Olap(ctx context.Context, req *OlapRequest) (*OlapResponse, error) {
	var out OlapResponse
	if _, err := s.rest.PostV2(ctx, EndpointOlap, nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// OlapField describes one column from /olap/columns. This endpoint is the
// authoritative live field list — deprecated fields are simply not returned.
type OlapField struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"` // ENUM STRING ID DATETIME INTEGER PERCENT DURATION_IN_SECONDS AMOUNT MONEY
	AggregationAllow bool     `json:"aggregationAllowed"`
	GroupingAllow    bool     `json:"groupingAllowed"`
	FilteringAllow   bool     `json:"filteringAllowed"`
	Tags             []string `json:"tags"`
}

func (s *Service) OlapFields(ctx context.Context, rt ReportType) (map[string]OlapField, error) {
	out := map[string]OlapField{}
	q := url.Values{"reportType": {string(rt)}}
	if _, err := s.rest.GetV2(ctx, EndpointOlapColumns, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// StoreBalance is /v2/reports/balance/stores. Always prefer this over OLAP's
// StartBalance/FinalBalance aggregates: those sum the entire transaction table
// over all DB history with no optimization and can hang the server. The docs
// say to use this endpoint instead "in all cases" since 5.2.
type StoreBalance struct {
	Store   string  `json:"store"`
	Product string  `json:"product"`
	Amount  float64 `json:"amount"`
	Sum     float64 `json:"sum"`
}

func (s *Service) StoreBalances(ctx context.Context, at time.Time, stores, products []string) ([]StoreBalance, error) {
	q := url.Values{"timestamp": {at.Format(rest.Stamp)}}
	for _, s := range stores {
		q.Add("store", s)
	}
	for _, p := range products {
		q.Add("product", p)
	}
	var out []StoreBalance
	if _, err := s.rest.GetV2(ctx, EndpointBalanceStores, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}
