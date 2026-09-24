package recipes

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Mixing size strategies inside one chart hierarchy multiplies to zero: nothing
// is written off at all, and iiko reports no error. The root dish decides the
// mode, so a chart that disagrees with it is refused before it is saved.
func TestSaveRejectsAMixedSizeStrategy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		chart   AssemblyChartDto
		root    ProductSizeAssemblyStrategy
		wantErr string
	}{
		{
			name:  "matching COMMON is saved",
			chart: AssemblyChartDto{AssembledProductID: "p-1", ProductSizeAssemblyStrategy: "COMMON"},
			root:  "COMMON",
		},
		{
			name:  "matching SPECIFIC is saved",
			chart: AssemblyChartDto{AssembledProductID: "p-1", ProductSizeAssemblyStrategy: "SPECIFIC"},
			root:  "SPECIFIC",
		},
		{
			name:    "a chart disagreeing with the root is refused",
			chart:   AssemblyChartDto{AssembledProductID: "p-1", ProductSizeAssemblyStrategy: "SPECIFIC"},
			root:    "COMMON",
			wantErr: "no write-off happens at all",
		},
		{
			name:    "no product",
			chart:   AssemblyChartDto{ProductSizeAssemblyStrategy: "COMMON"},
			root:    "COMMON",
			wantErr: "assembledProductId is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = w.Write([]byte(`{"result":"SUCCESS","response":{"id":"c-1"}}`))
			})
			_, err := s.SaveChart(context.Background(), tt.chart, tt.root)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("save: %v", err)
			}
			if !strings.HasSuffix(gotPath, EndpointChartsSave) {
				t.Errorf("path = %q", gotPath)
			}
		})
	}
}

func TestDeleteAndHistoryHitTheirPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		call     func(*Service) error
		wantPath string
		wantQ    map[string]string
	}{
		{
			name: "delete", wantPath: EndpointChartsDelete,
			call: func(s *Service) error { return s.DeleteChart(context.Background(), "c-1") },
		},
		{
			name: "history needs a product", wantPath: EndpointChartsGetHistory,
			wantQ: map[string]string{"productId": "p-1", "departmentId": "d-1"},
			call: func(s *Service) error {
				_, err := s.ChartHistory(context.Background(), "p-1", "d-1")
				return err
			},
		},
		{
			name: "history without a department omits it", wantPath: EndpointChartsGetHistory,
			wantQ: map[string]string{"productId": "p-1"},
			call: func(s *Service) error {
				_, err := s.ChartHistory(context.Background(), "p-1", "")
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
				_, _ = w.Write([]byte(`{"knownRevision":-1,"assemblyCharts":[],"preparedCharts":[]}`))
			})
			if err := tt.call(s); err != nil {
				t.Fatalf("call: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.wantPath) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.wantPath)
			}
			for k, want := range tt.wantQ {
				if gotQ.Get(k) != want {
					t.Errorf("%s = %q, want %q", k, gotQ.Get(k), want)
				}
			}
			if _, present := gotQ["departmentId"]; present && tt.wantQ["departmentId"] == "" {
				t.Error("an empty departmentId must be omitted, not sent blank")
			}
		})
	}
}

func TestChartWritesRejectMissingIdentifiers(t *testing.T) {
	t.Parallel()
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not reach iiko without an identifier")
		w.WriteHeader(http.StatusInternalServerError)
	})
	if err := s.DeleteChart(context.Background(), " "); err == nil {
		t.Error("delete must require a chart id")
	}
	if _, err := s.ChartHistory(context.Background(), "", "d-1"); err == nil {
		t.Error("history must require a productId")
	}
}
