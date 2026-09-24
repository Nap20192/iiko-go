//go:build integration

package iikoserver_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver"
	"github.com/Nap20192/iiko-go/iikoserver/corporation"
	"github.com/Nap20192/iiko-go/iikoserver/reports"
)

// Integration test against a real iiko stand. Behind the `integration` build
// tag so `go test ./...` stays hermetic and fast:
//
//	go test -tags integration ./pkg/iikoserver/ -run TestLive -v
//
//	IIKO_BASE_URL=https://your-stand.iiko.it/resto \
//	IIKO_LOGIN=... IIKO_PASSWORD=... IIKO_SKIP_TLS_VERIFY=1 \
//	  go test ./pkg/iikoserver/ -run TestLive -v
//
// Every response is dumped to testdata/live/ so real payloads can be turned
// into unit-test fixtures afterwards — which is how the guesses in this package
// (field names, envelope shapes, enum spellings) get confirmed or corrected.
// Pattern borrowed from Nap20192/bakery.
//
// A free demo stand is available from iiko on request (api@iiko.ru).
func TestLiveAgainstRealStand(t *testing.T) {
	base := os.Getenv("IIKO_BASE_URL")
	login := os.Getenv("IIKO_LOGIN")
	pass := os.Getenv("IIKO_PASSWORD")
	if base == "" || login == "" || pass == "" {
		t.Skip("set IIKO_BASE_URL, IIKO_LOGIN and IIKO_PASSWORD to run live tests")
	}

	c := iikoserver.New(iikoserver.Config{
		BaseURL:  base,
		Login:    login,
		Password: pass,
		SkipTLS:  os.Getenv("IIKO_SKIP_TLS_VERIFY") == "1",
		Timeout:  3 * time.Minute,
	})
	ctx := context.Background()

	// Releasing the licence seat matters even on a demo stand.
	t.Cleanup(func() {
		if err := c.Close(context.Background()); err != nil {
			t.Logf("logout failed, seat may stay held: %v", err)
		}
	})

	dir := t.TempDir()
	if keep := os.Getenv("IIKO_FIXTURE_DIR"); keep != "" {
		if err := os.MkdirAll(keep, 0o750); err != nil {
			t.Fatalf("fixture dir: %v", err)
		}
		dir = keep
	}
	dump := func(name string, v any) {
		t.Helper()
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Errorf("marshal %s: %v", name, err)
			return
		}
		if err := os.WriteFile(filepath.Join(dir, name), append(data, '\n'), 0o600); err != nil {
			t.Errorf("write %s: %v", name, err)
		}
	}

	if st, err := c.Corporation.ServerType(ctx); err != nil {
		t.Logf("serverType unavailable (expected on some standalone RMS): %v", err)
	} else {
		t.Logf("server type: %s", st)
		dump("00_server_type.json", map[string]string{"serverType": st})
	}

	deps, err := c.Corporation.ListEntities(ctx, corporation.KindDepartments, false)
	if err != nil {
		t.Fatalf("departments: %v", err)
	}
	dump("01_departments.json", deps)
	t.Logf("departments: %d", len(deps))

	// OLAP columns is the authoritative field vocabulary; capturing it tells us
	// whether the field names hardcoded in the tool enums actually exist here.
	fields, err := c.Reports.OlapFields(ctx, reports.ReportSales)
	if err != nil {
		t.Errorf("olap columns: %v", err)
	} else {
		dump("02_olap_columns_sales.json", fields)
		t.Logf("SALES fields: %d", len(fields))
		for _, want := range []string{"OpenDate.Typed", "DishName", "DishDiscountSumInt", "Department"} {
			if _, ok := fields[want]; !ok {
				t.Errorf("field %q is offered by this server's tool enum but does not exist here", want)
			}
		}
	}

	to := time.Now().Truncate(24 * time.Hour)
	from := to.AddDate(0, 0, -7)
	req, err := reports.NewOlapRequest(reports.ReportSales, from, to, []string{"OpenDate.Typed"}, nil, []string{"DishDiscountSumInt"})
	if err != nil {
		t.Fatalf("build olap: %v", err)
	}
	if resp, err := c.Reports.Olap(ctx, req); err != nil {
		t.Errorf("olap: %v", err)
	} else {
		dump("03_olap_sales_week.json", resp)
		t.Logf("olap rows: %d", len(resp.Data))
	}

	if charts, err := c.Recipes.ChartsGetAll(ctx, to, to, false, false); err != nil {
		t.Logf("assemblyCharts/getAll: %v", err)
	} else {
		dump("04_assembly_charts.json", charts)
		t.Logf("charts: %d, knownRevision=%d (usable for incremental sync: %v)",
			len(charts.AssemblyCharts), charts.KnownRevision, charts.UsableForIncrementalSync())
		// Confirm the enum spelling the docs and bakery disagree on.
		for _, ch := range charts.AssemblyCharts {
			if ch.ProductSizeAssemblyStrategy != "" {
				t.Logf("observed productSizeAssemblyStrategy: %q", ch.ProductSizeAssemblyStrategy)
				break
			}
		}
	}

	t.Logf("fixtures written to %s", dir)
}
