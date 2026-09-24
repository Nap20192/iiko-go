package recipes

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the technological charts half of the iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// Technological charts (техкарты) — recipes. iiko 6.0+.
//
// Structure adapted from Nap20192/bakery's internal/outbound/iiko, which models
// this domain well. Corrections applied: context plumbing, envelope-tolerant
// decoding, and the size-strategy enum (see WriteoffStrategy below).

// ProductWriteoffStrategy decides whether selling the dish writes off its
// ingredients or the finished product itself.
type ProductWriteoffStrategy string

const (
	WriteoffAssemble ProductWriteoffStrategy = "ASSEMBLE" // write off ingredients
	WriteoffDirectly ProductWriteoffStrategy = "DIRECTLY" // write off the finished item
)

// ProductSizeAssemblyStrategy is a genuine landmine. The mode is decided by the
// ROOT dish, not by each intermediate prep, so mixing modes inside one chart
// hierarchy multiplies to zero and NO write-off happens at all.
//
// The docs say, verbatim, `Enum: "COMMON", "SPECIFIC"`, and SEPARATE does not
// appear anywhere in the 150-page corpus — so Nap20192/bakery's SEPARATE has no
// documented basis and its origin is unknown. SPECIFIC is what we emit.
// SEPARATE is still accepted on input: it costs three lines, and if some build
// really does return it, the failure mode is a silent zero write-off.
type ProductSizeAssemblyStrategy string

const (
	SizeAssemblyCommon   ProductSizeAssemblyStrategy = "COMMON"
	SizeAssemblySpecific ProductSizeAssemblyStrategy = "SPECIFIC"
	SizeAssemblySeparate ProductSizeAssemblyStrategy = "SEPARATE" // undocumented; accepted defensively
)

func (s ProductSizeAssemblyStrategy) PerSize() bool {
	return s == SizeAssemblySpecific || s == SizeAssemblySeparate
}

// StoreSpecification scopes a chart line to a set of departments.
// {departments: [], inverse: true} means every department INCLUDING future
// ones — charts carry no versioning of department scope.
type StoreSpecification struct {
	Departments []string `json:"departments"`
	Inverse     bool     `json:"inverse"`
}

func (s StoreSpecification) Describe() string {
	switch {
	case len(s.Departments) == 0 && s.Inverse:
		return "all departments (including future)"
	case len(s.Departments) == 0:
		return "none"
	case s.Inverse:
		return fmt.Sprintf("all except %d department(s)", len(s.Departments))
	default:
		return fmt.Sprintf("%d department(s)", len(s.Departments))
	}
}

type SizeSpecification struct {
	SizeID string `json:"sizeId"`
}

// AssemblyChartItem is one line of a source recipe.
//
// Units trap: AmountIn/Middle/Out are in the ingredient's MAIN units, not the
// kilograms iikoOffice renders. iiko computes in product units and derives the
// packaging/kg columns for display only. AmountIn is what drives write-off.
type AssemblyChartItem struct {
	ID         string  `json:"id"`
	SortWeight float64 `json:"sortWeight"`

	ProductID                string              `json:"productId"`
	ProductSizeSpecification *SizeSpecification  `json:"productSizeSpecification"` // nil = common
	StoreSpecification       *StoreSpecification `json:"storeSpecification"`       // nil = all departments

	AmountIn     float64 `json:"amountIn"`     // gross
	AmountMiddle float64 `json:"amountMiddle"` // after cold processing
	AmountOut    float64 `json:"amountOut"`    // net / yield

	AmountIn1  float64 `json:"amountIn1"`
	AmountOut1 float64 `json:"amountOut1"`
	AmountIn2  float64 `json:"amountIn2"`
	AmountOut2 float64 `json:"amountOut2"`
	AmountIn3  float64 `json:"amountIn3"`
	AmountOut3 float64 `json:"amountOut3"`

	PackageCount  float64 `json:"packageCount"`
	PackageTypeID *string `json:"packageTypeId"`
}

type AssemblyChartDto struct {
	ID                 string  `json:"id"`
	AssembledProductID string  `json:"assembledProductId"`
	DateFrom           string  `json:"dateFrom"`
	DateTo             *string `json:"dateTo"` // nil = open-ended

	AssembledAmount float64 `json:"assembledAmount"`

	ProductWriteoffStrategy                   ProductWriteoffStrategy     `json:"productWriteoffStrategy"`
	EffectiveDirectWriteoffStoreSpecification StoreSpecification          `json:"effectiveDirectWriteoffStoreSpecification"`
	ProductSizeAssemblyStrategy               ProductSizeAssemblyStrategy `json:"productSizeAssemblyStrategy"`

	Items []AssemblyChartItem `json:"items"`

	TechnologyDescription string `json:"technologyDescription"`
	Description           string `json:"description"`
	Appearance            string `json:"appearance"`
	Organoleptic          string `json:"organoleptic"`
	OutputComment         string `json:"outputComment"`
}

// PreparedChartItem is a line of a recipe already decomposed to final
// ingredients — preps expanded, nothing left to resolve.
type PreparedChartItem struct {
	ID         string  `json:"id"`
	SortWeight float64 `json:"sortWeight"`

	ProductID                string              `json:"productId"`
	ProductSizeSpecification *SizeSpecification  `json:"productSizeSpecification"`
	StoreSpecification       *StoreSpecification `json:"storeSpecification"`

	Amount float64 `json:"amount"`
}

type PreparedChartDto struct {
	ID                 string  `json:"id"`
	AssembledProductID string  `json:"assembledProductId"`
	DateFrom           string  `json:"dateFrom"`
	DateTo             *string `json:"dateTo"`

	EffectiveDirectWriteoffStoreSpecification StoreSpecification          `json:"effectiveDirectWriteoffStoreSpecification"`
	ProductSizeAssemblyStrategy               ProductSizeAssemblyStrategy `json:"productSizeAssemblyStrategy"`

	Items []PreparedChartItem `json:"items"`
}

// ChartResultDto is the shared response of every assemblyCharts method.
type ChartResultDto struct {
	// KnownRevision feeds getAllUpdate. It is -1 on every method except
	// getAll/getAllUpdate, meaning that response cannot seed incremental sync.
	KnownRevision int `json:"knownRevision"`

	AssemblyCharts []AssemblyChartDto `json:"assemblyCharts"`
	PreparedCharts []PreparedChartDto `json:"preparedCharts"`

	// getAllUpdate only.
	DeletedAssemblyChartIDs []string `json:"deletedAssemblyChartIds"`
	DeletedPreparedChartIDs []string `json:"deletedPreparedChartIds"`
}

// UsableForIncrementalSync reports whether this response can seed getAllUpdate.
func (r *ChartResultDto) UsableForIncrementalSync() bool { return r.KnownRevision >= 0 }

// ChartsGetAll returns every chart effective in the date window.
// includePrepared defaults false upstream and is expensive; ask for it only
// when you actually need decomposed recipes.
func (s *Service) ChartsGetAll(ctx context.Context, from, to time.Time, includeDeleted, includePrepared bool) (*ChartResultDto, error) {
	q := url.Values{
		"dateFrom":               {from.Format(rest.QueryV2)},
		"dateTo":                 {to.Format(rest.QueryV2)},
		"includeDeletedProducts": {rest.BoolStr(includeDeleted)},
		"includePreparedCharts":  {rest.BoolStr(includePrepared)},
	}
	var out ChartResultDto
	if _, err := s.rest.GetV2(ctx, EndpointChartsGetAll, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ChartPrepared returns the recipe decomposed to final ingredients — the shape
// you want for consumption and costing, since preps are already expanded.
func (s *Service) ChartPrepared(ctx context.Context, date time.Time, productID, departmentID string) (*ChartResultDto, error) {
	return s.chartFor(ctx, EndpointChartsGetPrepared, date, productID, departmentID)
}

// ChartAssembled returns the source recipe, first level only: preps appear as
// ingredients rather than being expanded.
func (s *Service) ChartAssembled(ctx context.Context, date time.Time, productID, departmentID string) (*ChartResultDto, error) {
	return s.chartFor(ctx, EndpointChartsGetAssembled, date, productID, departmentID)
}

func (s *Service) chartFor(ctx context.Context, endpoint string, date time.Time, productID, departmentID string) (*ChartResultDto, error) {
	if productID == "" {
		return nil, fmt.Errorf("product_id is required. Find it with find_products")
	}
	q := url.Values{"date": {date.Format(rest.QueryV2)}, "productId": {productID}}
	if departmentID != "" {
		q.Set("departmentId", departmentID)
	}
	var out ChartResultDto
	if _, err := s.rest.GetV2(ctx, endpoint, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ChartByID fetches one chart by its own id.
func (s *Service) ChartByID(ctx context.Context, id string) (*ChartResultDto, error) {
	if id == "" {
		return nil, fmt.Errorf("chart id is required")
	}
	var out ChartResultDto
	if _, err := s.rest.GetV2(ctx, EndpointChartsByID, url.Values{"id": {id}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
