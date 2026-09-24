package reports

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// v1 reports take DD.MM.YYYY while everything under /v2/ takes ISO. One global
// formatter here is the single most likely way to silently read the wrong month.
func TestV1ReportsUseTheDottedDateFormat(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	to, _ := rest.ParseDay("2026-03-31")

	tests := []struct {
		name     string
		call     func(*Service) error
		wantPath string
	}{
		{
			name: "sales", wantPath: EndpointSalesReport,
			call: func(s *Service) error {
				_, err := s.SalesReport(context.Background(), from, to, false, false)
				return err
			},
		},
		{
			name: "productExpense", wantPath: EndpointProductExpense,
			call: func(s *Service) error {
				_, err := s.ProductExpense(context.Background(), from, to, -1, -1)
				return err
			},
		},
		{
			name: "monthlyIncomePlan", wantPath: EndpointMonthlyIncomePlan,
			call: func(s *Service) error {
				_, err := s.MonthlyIncomePlan(context.Background(), from, to)
				return err
			},
		},
		{
			name: "storeOperations", wantPath: EndpointStoreOperations,
			call: func(s *Service) error {
				_, err := s.StoreOperations(context.Background(), StoreOperationsFilter{From: from, To: to})
				return err
			},
		},
		{
			name: "olap v1", wantPath: EndpointOlapV1,
			call: func(s *Service) error {
				_, err := s.OlapV1(context.Background(), OlapV1Request{
					Report: "STOCK", From: from, To: to,
				})
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotQ url.Values
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotQ, gotPath = r.URL.Query(), r.URL.Path
				_, _ = w.Write([]byte(`<r></r>`))
			})
			if err := tt.call(s); err != nil {
				t.Fatalf("call: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.wantPath) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.wantPath)
			}
			gotFrom := gotQ.Get("dateFrom") + gotQ.Get("from")
			gotTo := gotQ.Get("dateTo") + gotQ.Get("to")
			if gotFrom != "01.03.2026" || gotTo != "31.03.2026" {
				t.Errorf("v1 dates must be DD.MM.YYYY, got from=%q to=%q", gotFrom, gotTo)
			}
		})
	}
}

// STOCK exists only on OLAP v1; v2 has three report types and no fourth.
func TestOlapV1AcceptsStockAndV2DoesNot(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	to, _ := rest.ParseDay("2026-03-31")

	var gotQ url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		_, _ = w.Write([]byte(`<r></r>`))
	})
	if _, err := s.OlapV1(context.Background(), OlapV1Request{Report: "STOCK", From: from, To: to}); err != nil {
		t.Fatalf("STOCK is a valid v1 report: %v", err)
	}
	if gotQ.Get("report") != "STOCK" {
		t.Errorf("report = %q", gotQ.Get("report"))
	}
	// buildSummary/summary can hang the server on a large chain, so it stays off
	// unless asked and is sent explicitly either way.
	if gotQ.Get("summary") != "false" {
		t.Errorf("summary must be sent explicitly as false, got %q", gotQ.Get("summary"))
	}
	if _, err := NewOlapRequest(ReportType("STOCK"), from, to, nil, nil, []string{"x"}); err == nil {
		t.Error("v2 has no STOCK report type and must reject it")
	}
}

// The delivery reports answer XML even though the docs show JSON samples, and
// their department filter is a literal-brace token that has to survive encoding.
func TestDeliveryReportsEncodeTheBraceDepartmentFilter(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	to, _ := rest.ParseDay("2026-03-02")

	var gotRawQuery, gotQ string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		gotQ = r.URL.Query().Get("department")
		// XML, despite the docs' JSON-labelled code blocks.
		_, _ = w.Write([]byte(`<deliveryReport><row><count>3</count></row></deliveryReport>`))
	})
	if _, err := s.DeliveryReport(context.Background(), DeliveryCouriers,
		DeliveryFilter{From: from, To: to, DepartmentCode: "5"}); err != nil {
		t.Fatalf("delivery: %v", err)
	}
	if gotQ != `{code="5"}` {
		t.Errorf("department must arrive as the literal brace token, got %q", gotQ)
	}
	if strings.Contains(gotRawQuery, `{code="5"}`) {
		t.Errorf("the brace token must be percent-encoded on the wire, raw query was %q", gotRawQuery)
	}
}

func TestDeliveryReportRejectsAnUnknownKind(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not reach iiko for an unknown report kind")
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := s.DeliveryReport(context.Background(), DeliveryKind("nope"),
		DeliveryFilter{From: from, To: from}); err == nil {
		t.Error("an unknown delivery report kind must be rejected")
	}
}

// presetId overrides every other filter except the dates, so sending both is a
// silent lie about what was asked for.
func TestStoreOperationsWarnsWhenPresetOverridesFilters(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not send filters that the preset would silently override")
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := s.StoreOperations(context.Background(), StoreOperationsFilter{
		From: from, To: from, PresetID: "p-1", Stores: []string{"s-1"},
	})
	if err == nil || !strings.Contains(err.Error(), "preset_id overrides") {
		t.Errorf("want a preset-overrides error, got %v", err)
	}
}

// showCostCorrections is honoured only alongside documentTypes; on its own it is
// silently dropped, which reads as "corrections were zero".
func TestStoreOperationsRejectsCostCorrectionsWithoutDocumentTypes(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not send a flag iiko will silently ignore")
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := s.StoreOperations(context.Background(), StoreOperationsFilter{
		From: from, To: from, ShowCostCorrections: true,
	})
	if err == nil || !strings.Contains(err.Error(), "documentTypes") {
		t.Errorf("want an error naming documentTypes, got %v", err)
	}
}

// productArticle wins over product when both are given, so sending both hides
// which one actually selected the rows.
func TestIngredientEntryRejectsBothProductSelectors(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not send two selectors when only one takes effect")
		w.WriteHeader(http.StatusInternalServerError)
	})
	_, err := s.IngredientEntry(context.Background(), from, from, "p-1", "ART-1", false)
	if err == nil || !strings.Contains(err.Error(), "takes priority") {
		t.Errorf("want an error explaining the precedence, got %v", err)
	}
}
