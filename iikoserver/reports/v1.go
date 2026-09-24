package reports

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// The v1 report family is XML and dates it DD.MM.YYYY, unlike everything under
// /v2/. rest.ReportV1 is that layout; there is no global date format here.

// rawReport is the response shape of the v1 reports: iiko returns a report-
// specific XML tree, and the useful part is different per report, so the raw
// document is handed back rather than guessed at.
type rawReport struct {
	XML string
}

func (s *Service) getV1(ctx context.Context, path string, q url.Values) (*rawReport, error) {
	data, err := s.rest.Do(ctx, "GET", path, q, nil, "")
	if err != nil {
		return nil, err
	}
	return &rawReport{XML: string(data)}, nil
}

func v1Window(from, to time.Time) url.Values {
	return url.Values{
		"dateFrom": {from.Format(rest.ReportV1)},
		"dateTo":   {to.Format(rest.ReportV1)},
	}
}

// SalesReport is the v1 revenue report.
func (s *Service) SalesReport(ctx context.Context, from, to time.Time, dishDetails, allRevenue bool) (*rawReport, error) {
	q := v1Window(from, to)
	q.Set("dishDetails", rest.BoolStr(dishDetails))
	q.Set("allRevenue", rest.BoolStr(allRevenue))
	return s.getV1(ctx, EndpointSalesReport, q)
}

// ProductExpense reports consumption per product. hourFrom and hourTo default to
// -1, meaning the whole day; they are sent explicitly.
func (s *Service) ProductExpense(ctx context.Context, from, to time.Time, hourFrom, hourTo int) (*rawReport, error) {
	q := v1Window(from, to)
	q.Set("hourFrom", fmt.Sprint(hourFrom))
	q.Set("hourTo", fmt.Sprint(hourTo))
	return s.getV1(ctx, EndpointProductExpense, q)
}

// MonthlyIncomePlan reports plan against actual.
func (s *Service) MonthlyIncomePlan(ctx context.Context, from, to time.Time) (*rawReport, error) {
	return s.getV1(ctx, EndpointMonthlyIncomePlan, v1Window(from, to))
}

// IngredientEntry traces where an ingredient entered stock.
//
// productArticle takes priority over product when both are set, so passing both
// hides which one selected the rows; this refuses instead.
func (s *Service) IngredientEntry(ctx context.Context, from, to time.Time, productID, productArticle string, includeSubtree bool) (*rawReport, error) {
	if productID != "" && productArticle != "" {
		return nil, fmt.Errorf("pass product or productArticle, not both: productArticle takes priority and the other would be silently ignored")
	}
	if productID == "" && productArticle == "" {
		return nil, fmt.Errorf("product or productArticle is required")
	}
	q := v1Window(from, to)
	if productID != "" {
		q.Set("product", productID)
	} else {
		q.Set("productArticle", productArticle)
	}
	q.Set("includeSubtree", rest.BoolStr(includeSubtree))
	return s.getV1(ctx, EndpointIngredientEntry, q)
}

// StoreOperationsFilter selects stock movements.
//
// PresetID overrides every other filter except the dates, and ShowCostCorrections
// is honoured only alongside DocumentTypes. Both are refused rather than sent
// into silence.
type StoreOperationsFilter struct {
	From, To            time.Time
	Stores              []string
	DocumentTypes       []string
	ProductDetalization bool
	ShowCostCorrections bool
	PresetID            string
}

// StoreOperations reports stock movements for a period.
func (s *Service) StoreOperations(ctx context.Context, f StoreOperationsFilter) (*rawReport, error) {
	if f.PresetID != "" && (len(f.Stores) > 0 || len(f.DocumentTypes) > 0 || f.ProductDetalization || f.ShowCostCorrections) {
		return nil, fmt.Errorf("preset_id overrides every filter except the dates; send the preset alone or drop it")
	}
	if f.ShowCostCorrections && len(f.DocumentTypes) == 0 {
		return nil, fmt.Errorf("showCostCorrections is honoured only together with documentTypes; on its own iiko ignores it and the report reads as if there were no corrections")
	}
	q := v1Window(f.From, f.To)
	if f.PresetID != "" {
		q.Set("presetId", f.PresetID)
		return s.getV1(ctx, EndpointStoreOperations, q)
	}
	for _, st := range f.Stores {
		q.Add("stores", st)
	}
	for _, dt := range f.DocumentTypes {
		q.Add("documentTypes", dt)
	}
	q.Set("productDetalization", rest.BoolStr(f.ProductDetalization))
	if len(f.DocumentTypes) > 0 {
		q.Set("showCostCorrections", rest.BoolStr(f.ShowCostCorrections))
	}
	return s.getV1(ctx, EndpointStoreOperations, q)
}

// StoreReportPresets lists the saved store-report configurations.
func (s *Service) StoreReportPresets(ctx context.Context) (*rawReport, error) {
	return s.getV1(ctx, EndpointStoreReportPresets, nil)
}

// OlapV1Request is the legacy OLAP call. It is the only place STOCK exists:
// OLAP v2 has three report types and no fourth.
type OlapV1Request struct {
	Report   string // SALES | TRANSACTIONS | DELIVERIES | STOCK
	From, To time.Time
	GroupRow []string
	GroupCol []string
	Agg      []string
	Summary  bool // default false; true can hang the server on a large chain
}

// OlapV1 runs the legacy OLAP report.
func (s *Service) OlapV1(ctx context.Context, req OlapV1Request) (*rawReport, error) {
	if strings.TrimSpace(req.Report) == "" {
		return nil, fmt.Errorf("report is required: SALES, TRANSACTIONS, DELIVERIES or STOCK")
	}
	q := url.Values{
		"report":  {req.Report},
		"from":    {req.From.Format(rest.ReportV1)},
		"to":      {req.To.Format(rest.ReportV1)},
		"summary": {rest.BoolStr(req.Summary)},
	}
	for _, g := range req.GroupRow {
		q.Add("groupRow", g)
	}
	for _, g := range req.GroupCol {
		q.Add("groupCol", g)
	}
	for _, a := range req.Agg {
		q.Add("agr", a)
	}
	return s.getV1(ctx, EndpointOlapV1, q)
}

// DeliveryKind selects one of the six delivery reports.
type DeliveryKind string

const (
	DeliveryConsolidated     DeliveryKind = "consolidated"
	DeliveryCouriers         DeliveryKind = "couriers"
	DeliveryOrderCycle       DeliveryKind = "orderCycle"
	DeliveryHalfHourDetailed DeliveryKind = "halfHourDetailed"
	DeliveryRegions          DeliveryKind = "regions"
	DeliveryLoyalty          DeliveryKind = "loyalty"
)

func deliveryEndpoint(k DeliveryKind) string {
	switch k {
	case DeliveryConsolidated:
		return EndpointDeliveryConsolidated
	case DeliveryCouriers:
		return EndpointDeliveryCouriers
	case DeliveryOrderCycle:
		return EndpointDeliveryOrderCycle
	case DeliveryHalfHourDetailed:
		return EndpointDeliveryHalfHourDetailed
	case DeliveryRegions:
		return EndpointDeliveryRegions
	case DeliveryLoyalty:
		return EndpointDeliveryLoyalty
	default:
		return ""
	}
}

// DeliveryFilter scopes a delivery report.
//
// DepartmentCode goes out as the literal token {code="..."}, which is what these
// endpoints expect; net/url percent-encodes it on the wire.
type DeliveryFilter struct {
	From, To         time.Time
	DepartmentCode   string
	WriteoffAccounts []string
	MetricType       string // loyalty only: AVERAGE | MINIMUM | MAXIMUM
}

// DeliveryReport runs one of the six delivery reports.
//
// These answer XML even though the docs show JSON samples, so the raw document
// is returned rather than a decoded shape.
func (s *Service) DeliveryReport(ctx context.Context, kind DeliveryKind, f DeliveryFilter) (*rawReport, error) {
	endpoint := deliveryEndpoint(kind)
	if endpoint == "" {
		return nil, fmt.Errorf("unknown delivery report %q; valid: consolidated, couriers, orderCycle, halfHourDetailed, regions, loyalty", kind)
	}
	q := v1Window(f.From, f.To)
	if f.DepartmentCode != "" {
		q.Set("department", `{code="`+f.DepartmentCode+`"}`)
	}
	for _, a := range f.WriteoffAccounts {
		q.Add("writeoffAccounts", a)
	}
	if f.MetricType != "" {
		q.Set("metricType", f.MetricType)
	}
	return s.getV1(ctx, endpoint, q)
}
