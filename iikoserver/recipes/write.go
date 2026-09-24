package recipes

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// SaveChart creates or replaces a technological chart.
//
// This changes what iiko writes off when the dish is sold, so it is a write in
// the fullest sense and belongs behind the server's write flag.
//
// rootStrategy is the size strategy of the ROOT dish of the hierarchy this chart
// belongs to. iiko decides which chart rows apply from the root, not per chart,
// and a chart in the other mode multiplies out to zero: nothing is written off
// and no error is reported. Passing the root's strategy lets that be caught here.
func (s *Service) SaveChart(ctx context.Context, chart AssemblyChartDto, rootStrategy ProductSizeAssemblyStrategy) (*ChartResultDto, error) {
	if strings.TrimSpace(chart.AssembledProductID) == "" {
		return nil, fmt.Errorf("assembledProductId is required: a chart with no dish writes nothing off")
	}
	if rootStrategy != "" && chart.ProductSizeAssemblyStrategy != "" &&
		chart.ProductSizeAssemblyStrategy != rootStrategy {
		return nil, fmt.Errorf(
			"this chart is %s but the root dish of its hierarchy is %s; iiko takes the mode from the root, "+
				"so the two multiply out and no write-off happens at all — and iiko reports no error",
			chart.ProductSizeAssemblyStrategy, rootStrategy)
	}
	var out ChartResultDto
	if _, err := s.rest.PostV2(ctx, EndpointChartsSave, nil, chart, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteChart removes a technological chart. Also a write: the dish stops being
// written off by recipe from the moment this lands.
func (s *Service) DeleteChart(ctx context.Context, chartID string) error {
	if strings.TrimSpace(chartID) == "" {
		return fmt.Errorf("chart id is required")
	}
	_, err := s.rest.PostV2(ctx, EndpointChartsDelete, url.Values{"id": {chartID}}, nil, nil)
	return err
}

// ChartHistory returns every version of a product's charts over time.
//
// productId is required; departmentId narrows it and is omitted when empty.
func (s *Service) ChartHistory(ctx context.Context, productID, departmentID string) (*ChartResultDto, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("productId is required")
	}
	q := url.Values{"productId": {productID}}
	if departmentID != "" {
		q.Set("departmentId", departmentID)
	}
	var out ChartResultDto
	if _, err := s.rest.GetV2(ctx, EndpointChartsGetHistory, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
