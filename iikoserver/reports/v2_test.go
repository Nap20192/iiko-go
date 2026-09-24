package reports

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// byPresetId's window is dateFrom inclusive, dateTo EXCLUSIVE, stamped to the
// second — and its toggle is `summary`, not the POST body's `buildSummary`.
func TestOlapByPresetIDUsesStampedDatesAndItsOwnSummaryName(t *testing.T) {
	t.Parallel()
	from, _ := rest.ParseDay("2026-03-01")
	to, _ := rest.ParseDay("2026-03-31")
	var gotQ url.Values
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ, gotPath = r.URL.Query(), r.URL.Path
		_, _ = w.Write([]byte(`[{"DishName":"Tea"}]`))
	})
	rows, err := s.OlapByPresetID(context.Background(), "p-1", from, to, false)
	if err != nil {
		t.Fatalf("byPresetId: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if !strings.HasSuffix(gotPath, "/api/v2/reports/olap/byPresetId/p-1") {
		t.Errorf("preset id belongs in the path, got %q", gotPath)
	}
	if gotQ.Get("dateFrom") != "2026-03-01T00:00:00" {
		t.Errorf("dateFrom must be stamped to the second, got %q", gotQ.Get("dateFrom"))
	}
	// the toggle is named `summary` here and `buildSummary` in the POST body
	if gotQ.Get("summary") != "false" {
		t.Errorf("summary must be sent explicitly as false, got %q", gotQ.Get("summary"))
	}
	if _, present := gotQ["buildSummary"]; present {
		t.Error("buildSummary belongs to the POST body, not to byPresetId")
	}
}

func TestOlapPresetsByTypeLowercasesTheType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, in, want string
		wantErr        bool
	}{
		{name: "stock", in: "stock", want: "/api/v2/reports/olap/presets/stock"},
		{name: "uppercase is normalised", in: "SALES", want: "/api/v2/reports/olap/presets/sales"},
		{name: "unknown type", in: "nonsense", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = w.Write([]byte(`[]`))
			})
			_, err := s.OlapPresetsByType(context.Background(), tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("an unknown preset type must be rejected")
				}
				return
			}
			if err != nil {
				t.Fatalf("presets: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.want) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.want)
			}
		})
	}
}

// Counteragent balances share the store-balance rule: a timestamp, not a date,
// and repeated filters rather than joined ones.
func TestCounteragentBalancesUseATimestamp(t *testing.T) {
	t.Parallel()
	at, _ := rest.ParseDay("2026-03-05")
	var gotQ url.Values
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ, gotPath = r.URL.Query(), r.URL.Path
		_, _ = w.Write([]byte(`[{"counteragent":"ООО Ромашка","sum":1200.5}]`))
	})
	got, err := s.CounteragentBalances(context.Background(), at, []string{"c1", "c2"}, []string{"a1"})
	if err != nil {
		t.Fatalf("balances: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 row, got %d", len(got))
	}
	if !strings.HasSuffix(gotPath, EndpointBalanceCounteragents) {
		t.Errorf("path = %q", gotPath)
	}
	if gotQ.Get("timestamp") != "2026-03-05T00:00:00" {
		t.Errorf("timestamp must be yyyy-MM-ddTHH:mm:ss, got %q", gotQ.Get("timestamp"))
	}
	if len(gotQ["counteragent"]) != 2 || len(gotQ["account"]) != 1 {
		t.Errorf("filters must repeat per value, got %v", gotQ)
	}
}

// EGAIS marks carry a sentinel far-future date meaning "not written off"; a
// client that reads it as a real date reports stock as disposed of in 2500.
func TestEgaisMarksExposeTheNotWrittenOffSentinel(t *testing.T) {
	t.Parallel()
	var gotQ url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		_, _ = w.Write([]byte(`{"fullUpdate":true,"marks":[
		  {"mark":"m-1","writeoffDate":"2500-01-01T00:00:00"},
		  {"mark":"m-2","writeoffDate":"2026-02-01T10:00:00"}]}`))
	})
	got, err := s.EgaisMarks(context.Background(), []string{"f1", "f2"}, -1)
	if err != nil {
		t.Fatalf("egais: %v", err)
	}
	if len(gotQ["fsRarId"]) != 2 {
		t.Errorf("fsRarId must repeat per value, got %v", gotQ["fsRarId"])
	}
	if !got.FullUpdate {
		t.Error("fullUpdate must be surfaced: it tells the caller to discard its cache")
	}
	if len(got.Marks) != 2 {
		t.Fatalf("want 2 marks, got %d", len(got.Marks))
	}
	if !got.Marks[0].NotWrittenOff() {
		t.Error("the 2500-01-01 sentinel means not written off, not a write-off in the year 2500")
	}
	if got.Marks[1].NotWrittenOff() {
		t.Error("a real write-off date must not read as the sentinel")
	}
}

func TestPresetListsHitTheirOwnPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		body     string
		call     func(*Service) error
		wantPath string
	}{
		{
			name: "olap presets", wantPath: EndpointOlapPresets,
			body: `[{"id":"p-1","name":"Daily","reportType":"SALES","groupByRowFields":["DishName"]}]`,
			call: func(s *Service) error {
				got, err := s.OlapPresets(context.Background())
				if err == nil && (len(got) != 1 || len(got[0].GroupByRowFields) != 1) {
					return fmt.Errorf("preset did not decode: %+v", got)
				}
				return err
			},
		},
		{
			name: "store report presets", wantPath: EndpointStoreReportPresets,
			body: `<presets><preset><id>sp-1</id></preset></presets>`,
			call: func(s *Service) error { _, err := s.StoreReportPresets(context.Background()); return err },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = w.Write([]byte(tt.body))
			})
			if err := tt.call(s); err != nil {
				t.Fatalf("call: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.wantPath) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.wantPath)
			}
		})
	}
}
