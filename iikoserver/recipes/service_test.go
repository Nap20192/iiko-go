package recipes

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

func TestSizeStrategyAcceptsBothSpellings(t *testing.T) {
	// Docs say SPECIFIC; a production client in the wild emits SEPARATE.
	// Getting this wrong means zero write-off, silently.
	if !SizeAssemblySpecific.PerSize() {
		t.Error("SPECIFIC must count as per-size")
	}
	if !SizeAssemblySeparate.PerSize() {
		t.Error("SEPARATE is the same thing under another name")
	}
	if SizeAssemblyCommon.PerSize() {
		t.Error("COMMON is not per-size")
	}
}
func TestStoreSpecificationDescribesInverseEmpty(t *testing.T) {
	// {departments: [], inverse: true} means everything, including departments
	// that do not exist yet — the single most misreadable value in the schema.
	all := StoreSpecification{Departments: nil, Inverse: true}
	if got := all.Describe(); !strings.Contains(got, "all departments") {
		t.Fatalf("empty+inverse means all departments, got %q", got)
	}
	none := StoreSpecification{}
	if got := none.Describe(); got != "none" {
		t.Fatalf("empty+not-inverse means none, got %q", got)
	}
}
func TestChartResultKnownRevisionGate(t *testing.T) {
	// knownRevision is -1 on every method except getAll/getAllUpdate; using
	// such a response to seed incremental sync silently loses updates.
	if (&ChartResultDto{KnownRevision: -1}).UsableForIncrementalSync() {
		t.Fatal("-1 must not be usable for getAllUpdate")
	}
	if !(&ChartResultDto{KnownRevision: 0}).UsableForIncrementalSync() {
		t.Fatal("0 is a valid revision")
	}
}

// The chart wrappers are thin — build a URL, call, decode — so one table covers
// all of them and asserts the thing that actually varies: which path is hit.
func TestChartEndpointsHitTheDocumentedPaths(t *testing.T) {
	t.Parallel()
	day, _ := rest.ParseDay("2026-03-05")
	const body = `{"knownRevision":7,"assemblyCharts":[{"id":"c1","assembledProductId":"p1"}],"preparedCharts":[]}`

	tests := []struct {
		name     string
		call     func(*Service) (*ChartResultDto, error)
		wantPath string
		wantQ    map[string]string
	}{
		{
			name: "getAll passes the window and both include flags",
			call: func(s *Service) (*ChartResultDto, error) {
				return s.ChartsGetAll(context.Background(), day, day, false, true)
			},
			wantPath: "/resto" + EndpointChartsGetAll,
			wantQ:    map[string]string{"dateFrom": "2026-03-05", "includeDeletedProducts": "false", "includePreparedCharts": "true"},
		},
		{
			name: "getAssembled is first-level only",
			call: func(s *Service) (*ChartResultDto, error) {
				return s.ChartAssembled(context.Background(), day, "p1", "d1")
			},
			wantPath: "/resto" + EndpointChartsGetAssembled,
			wantQ:    map[string]string{"date": "2026-03-05", "productId": "p1", "departmentId": "d1"},
		},
		{
			name:     "getPrepared decomposes to raw ingredients",
			call:     func(s *Service) (*ChartResultDto, error) { return s.ChartPrepared(context.Background(), day, "p1", "") },
			wantPath: "/resto" + EndpointChartsGetPrepared,
			wantQ:    map[string]string{"productId": "p1"},
		},
		{
			name:     "byId takes the chart's own id",
			call:     func(s *Service) (*ChartResultDto, error) { return s.ChartByID(context.Background(), "c1") },
			wantPath: "/resto" + EndpointChartsByID,
			wantQ:    map[string]string{"id": "c1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			var gotQ url.Values
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotQ = r.URL.Path, r.URL.Query()
				_, _ = w.Write([]byte(body))
			})
			res, err := tt.call(s)
			if err != nil {
				t.Fatalf("call: %v", err)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tt.wantPath)
			}
			for k, want := range tt.wantQ {
				if got := gotQ.Get(k); got != want {
					t.Errorf("query %s = %q, want %q", k, got, want)
				}
			}
			if !res.UsableForIncrementalSync() {
				t.Error("knownRevision 7 should be usable for getAllUpdate")
			}
		})
	}
}

// departmentId is optional; sending an empty one would filter to nothing.
func TestChartOmitsEmptyDepartment(t *testing.T) {
	t.Parallel()
	var q url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		_, _ = w.Write([]byte(`{"knownRevision":-1,"preparedCharts":[]}`))
	})
	if _, err := s.ChartPrepared(context.Background(), time.Now(), "p1", ""); err != nil {
		t.Fatal(err)
	}
	if _, present := q["departmentId"]; present {
		t.Error("empty departmentId must be omitted, not sent blank")
	}
}

func TestChartRequiresProductID(t *testing.T) {
	t.Parallel()
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) { t.Error("must not reach the server") })
	if _, err := s.ChartPrepared(context.Background(), time.Now(), "", ""); err == nil {
		t.Fatal("empty product_id must be rejected before the request")
	}
}
