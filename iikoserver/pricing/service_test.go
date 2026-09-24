package pricing

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

// newFake wires a Service to a stub iiko and captures the request it received.
func newFake(t *testing.T, body string) (*Service, *url.URL) {
	t.Helper()
	got := &url.URL{}
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		*got = *r.URL
		_, _ = w.Write([]byte(body))
	})
	return New(c), got
}

func day(s string) time.Time {
	t, err := time.Parse(rest.QueryV2, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestListPricesRequiresDateFrom(t *testing.T) {
	t.Parallel()

	// dateFrom is the one required parameter; without it iiko answers with an
	// error body rather than a price list, which decodes as an empty report.
	s, _ := newFake(t, `{"result":"SUCCESS","response":[]}`)
	_, _, err := s.ListPrices(context.Background(), time.Time{}, time.Time{}, nil, false, "", -1)
	if err == nil || !strings.Contains(err.Error(), "dateFrom") {
		t.Fatalf("got %v, want a dateFrom error", err)
	}
}

func TestListPricesDefaultsDateToTheDocumentedFarFuture(t *testing.T) {
	t.Parallel()

	s, got := newFake(t, `{"result":"SUCCESS","response":[],"revision":42}`)
	_, rev, err := s.ListPrices(context.Background(), day("2026-09-01"), time.Time{},
		[]string{"dep-1", "dep-2"}, true, PriceScheduled, 17)
	if err != nil {
		t.Fatal(err)
	}

	q := got.Query()
	// The docs' own default, not a guess: an omitted dateTo means 2500-01-01.
	for k, want := range map[string]string{
		"dateFrom":         "2026-09-01",
		"dateTo":           "2500-01-01",
		"includeOutOfSale": "true",
		"type":             "SCHEDULED",
		"revisionFrom":     "17",
	} {
		if got := q.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	// departmentId is repeatable: one key, several values, never a joined string.
	if deps := q["departmentId"]; len(deps) != 2 || deps[0] != "dep-1" || deps[1] != "dep-2" {
		t.Errorf("departmentId = %v", deps)
	}
	if rev == nil || *rev != 42 {
		t.Errorf("revision = %v, want 42: it is the cursor for the next call", rev)
	}
}

func TestListPricesRejectsAnUndocumentedType(t *testing.T) {
	t.Parallel()

	// BASE and SCHEDULED are the whole vocabulary; anything else is silently
	// ignored upstream and returns every price, which reads as a bug elsewhere.
	s, _ := newFake(t, `{"result":"SUCCESS","response":[]}`)
	_, _, err := s.ListPrices(context.Background(), day("2026-09-01"), time.Time{}, nil, false, "ALL", -1)
	if err == nil || !strings.Contains(err.Error(), "ALL") {
		t.Fatalf("got %v, want a rejected type", err)
	}
}

func TestListPricesDecodesScheduledPrices(t *testing.T) {
	t.Parallel()

	s, _ := newFake(t, `{"result":"SUCCESS","response":[
		{"departmentId":"dep-1","productId":"prod-1","prices":[
			{"dateFrom":"2026-09-01","dateTo":"2026-09-30","price":350.5,"documentId":"doc-1"}]}],
		"revision":7}`)
	prices, _, err := s.ListPrices(context.Background(), day("2026-09-01"), time.Time{}, nil, false, "", -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(prices) != 1 || len(prices[0].Prices) != 1 || prices[0].Prices[0].Price != 350.5 {
		t.Fatalf("got %+v", prices)
	}
}

func TestListMenuChangesRequiresBothDates(t *testing.T) {
	t.Parallel()

	s, _ := newFake(t, `{"result":"SUCCESS","response":[]}`)
	for _, tc := range []struct{ from, to time.Time }{
		{time.Time{}, day("2026-09-30")},
		{day("2026-09-01"), time.Time{}},
	} {
		if _, _, err := s.ListMenuChanges(context.Background(), tc.from, tc.to, "", -1); err == nil {
			t.Errorf("from=%v to=%v: want an error, both bounds are required", tc.from, tc.to)
		}
	}
}

func TestListMenuChangesSendsStatusOnlyWhenAsked(t *testing.T) {
	t.Parallel()

	s, got := newFake(t, `{"result":"SUCCESS","response":[],"revision":3}`)
	if _, _, err := s.ListMenuChanges(context.Background(), day("2026-09-01"), day("2026-09-30"), "", -1); err != nil {
		t.Fatal(err)
	}
	// An omitted status means every status upstream; sending an empty one does not.
	if _, ok := got.Query()["status"]; ok {
		t.Error("status must be absent, not empty")
	}
	if got.Query().Get("revisionFrom") != "-1" {
		t.Errorf("revisionFrom = %q", got.Query().Get("revisionFrom"))
	}

	if _, _, err := s.ListMenuChanges(context.Background(), day("2026-09-01"), day("2026-09-30"), StatusProcessed, -1); err != nil {
		t.Fatal(err)
	}
	if v := got.Query().Get("status"); v != "PROCESSED" {
		t.Errorf("status = %q", v)
	}
}

func TestListMenuChangesRejectsAnUndocumentedStatus(t *testing.T) {
	t.Parallel()

	s, _ := newFake(t, `{"result":"SUCCESS","response":[]}`)
	_, _, err := s.ListMenuChanges(context.Background(), day("2026-09-01"), day("2026-09-30"), "OPEN", -1)
	if err == nil || !strings.Contains(err.Error(), "OPEN") {
		t.Fatalf("got %v", err)
	}
}

func TestMenuChangeByIDDecodesABareObject(t *testing.T) {
	t.Parallel()

	// byId is not v2-wrapped, unlike the list on the same page.
	s, got := newFake(t, `{"id":"doc-1","documentNumber":"MC-7","status":"NEW","dateIncoming":"2026-09-01T00:00"}`)
	doc, err := s.MenuChangeByID(context.Background(), "doc-1")
	if err != nil {
		t.Fatal(err)
	}
	if doc.DocumentNumber != "MC-7" || doc.Status != StatusNew {
		t.Fatalf("got %+v", doc)
	}
	if got.Path != "/resto"+EndpointMenuChangeByID || got.Query().Get("id") != "doc-1" {
		t.Errorf("called %s?%s", got.Path, got.RawQuery)
	}
}

func TestMenuChangeByNumberDecodesABareArray(t *testing.T) {
	t.Parallel()

	s, got := newFake(t, `[{"id":"doc-1","documentNumber":"MC-7"},{"id":"doc-2","documentNumber":"MC-7"}]`)
	docs, err := s.MenuChangeByNumber(context.Background(), "MC-7")
	if err != nil {
		t.Fatal(err)
	}
	// A number is not unique across years, so this one really is an array.
	if len(docs) != 2 {
		t.Fatalf("got %+v", docs)
	}
	if got.Query().Get("documentNumber") != "MC-7" {
		t.Errorf("query was %s", got.RawQuery)
	}
}

func TestSaveMenuChangeTreatsResultErrorAsFailure(t *testing.T) {
	t.Parallel()

	// HTTP 200 with result=ERROR is how iiko rejects a write; decoding it as
	// success would report a saved order that does not exist.
	s, _ := newFake(t, `{"result":"ERROR","errors":[{"code":"E1","value":"dateIncoming is in the past"}]}`)
	_, err := s.SaveMenuChange(context.Background(), MenuChangeDocumentDto{DocumentNumber: "MC-7"})
	if err == nil || !strings.Contains(err.Error(), "dateIncoming is in the past") {
		t.Fatalf("got %v", err)
	}
}

func TestSaveMenuChangeDoesNotSendReadOnlyFields(t *testing.T) {
	t.Parallel()

	var body string
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		body = string(b)
		_, _ = w.Write([]byte(`{"result":"SUCCESS","response":{"id":"doc-1"}}`))
	})
	doc, err := New(c).SaveMenuChange(context.Background(), MenuChangeDocumentDto{
		DocumentNumber: "MC-7",
		ScheduleID:     "sched-1",
		Schedule:       PeriodScheduleDto{ID: "sched-1", Name: "Завтрак"},
	})
	if err != nil {
		t.Fatal(err)
	}
	// schedule is read-only; iiko rejects the document when it is echoed back.
	if strings.Contains(body, `"schedule"`) {
		t.Errorf("read-only schedule was sent: %s", body)
	}
	if !strings.Contains(body, `"scheduleId":"sched-1"`) {
		t.Errorf("scheduleId is the writable half and must be sent: %s", body)
	}
	if doc.ID != "doc-1" {
		t.Errorf("got %+v", doc)
	}
}

func TestSaveMenuChangeDropsTheReadOnlyLineNumber(t *testing.T) {
	t.Parallel()

	// num is assigned by iiko. Echoing it back on an edit renumbers nothing and
	// is rejected, so the generated marshaller drops it — pinned here because
	// the item list is the half of the document that actually sets prices.
	var body string
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		body = string(b)
		_, _ = w.Write([]byte(`{"result":"SUCCESS","response":{"id":"doc-1"}}`))
	})
	if _, err := New(c).SaveMenuChange(context.Background(), MenuChangeDocumentDto{
		Items: []MenuChangeDocumentItemDto{{Num: 3, ProductID: "prod-1", Price: 99}},
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(body, `"num"`) {
		t.Errorf("read-only num was sent: %s", body)
	}
	if !strings.Contains(body, `"price":99`) {
		t.Errorf("price must survive: %s", body)
	}
}

func TestListPriceCategoriesAlwaysSendsIncludeDeleted(t *testing.T) {
	t.Parallel()

	// Upstream defaults flip per endpoint, so it is never left to the server.
	s, got := newFake(t, `{"result":"SUCCESS","response":[],"revision":5}`)
	if _, _, err := s.ListPriceCategories(context.Background(), false, nil, -1); err != nil {
		t.Fatal(err)
	}
	if v := got.Query().Get("includeDeleted"); v != "false" {
		t.Errorf("includeDeleted = %q, want an explicit false", v)
	}
}

func TestListPeriodSchedulesRepeatsTheIDParameter(t *testing.T) {
	t.Parallel()

	s, got := newFake(t, `{"result":"SUCCESS","response":[],"revision":5}`)
	if _, _, err := s.ListPeriodSchedules(context.Background(), true, []string{"s-1", "s-2"}, 9); err != nil {
		t.Fatal(err)
	}
	if ids := got.Query()["id"]; len(ids) != 2 || ids[1] != "s-2" {
		t.Errorf("id = %v, want two separate values", ids)
	}
	if got.Query().Get("includeDeleted") != "true" || got.Query().Get("revisionFrom") != "9" {
		t.Errorf("query was %s", got.RawQuery)
	}
}

func TestByIDEndpointsDecodeBareObjects(t *testing.T) {
	t.Parallel()

	t.Run("priceCategory", func(t *testing.T) {
		t.Parallel()
		s, got := newFake(t, `{"id":"pc-1","name":"VIP","pricingStrategy":{"type":"PERCENT","percent":10}}`)
		cat, err := s.PriceCategoryByID(context.Background(), "pc-1")
		if err != nil {
			t.Fatal(err)
		}
		if cat.Name != "VIP" || cat.PricingStrategy.Percent != 10 {
			t.Fatalf("got %+v", cat)
		}
		if got.Path != "/resto"+EndpointPriceCategoryByID {
			t.Errorf("path %s", got.Path)
		}
	})
	t.Run("periodSchedule", func(t *testing.T) {
		t.Parallel()
		s, got := newFake(t, `{"id":"s-1","name":"Завтрак","deleted":false,
			"periods":[{"begin":"08:00","end":"11:00","daysOfWeek":[1,2,3,4,5]}]}`)
		sch, err := s.PeriodScheduleByID(context.Background(), "s-1")
		if err != nil {
			t.Fatal(err)
		}
		if len(sch.Periods) != 1 {
			t.Fatalf("got %+v", sch)
		}
		if got.Path != "/resto"+EndpointPeriodScheduleByID {
			t.Errorf("path %s", got.Path)
		}
	})
}

func TestDaysOfWeekAreMondayFirst(t *testing.T) {
	t.Parallel()

	// iiko numbers period days 1=Monday..7=Sunday. Two things nearby disagree:
	// java.util.Calendar is 1=Sunday, and this API's own quickLabels are 0..6.
	// Reading Sunday as Monday shifts a whole schedule by a day, silently.
	days, err := DaysOfWeek(PeriodScheduleItemDto{DaysOfWeek: []int{1, 5, 7}})
	if err != nil {
		t.Fatal(err)
	}
	want := []time.Weekday{time.Monday, time.Friday, time.Sunday}
	if len(days) != len(want) {
		t.Fatalf("got %v, want %v", days, want)
	}
	for i := range want {
		if days[i] != want[i] {
			t.Fatalf("got %v, want %v", days, want)
		}
	}
}

func TestDaysOfWeekRejectsTheQuickLabelsNumbering(t *testing.T) {
	t.Parallel()

	// 0 is valid in this API's quickLabels (0=Monday..6=Sunday) and invalid
	// here; accepting it would map Monday onto Sunday without complaining.
	if _, err := DaysOfWeek(PeriodScheduleItemDto{DaysOfWeek: []int{0}}); err == nil {
		t.Fatal("0 must be rejected: period days start at 1")
	}
	if _, err := DaysOfWeek(PeriodScheduleItemDto{DaysOfWeek: []int{8}}); err == nil {
		t.Fatal("8 must be rejected")
	}
	if days, err := DaysOfWeek(PeriodScheduleItemDto{DaysOfWeek: []int{7}}); err != nil ||
		len(days) != 1 || days[0] != time.Sunday {
		t.Fatalf("got %v, %v", days, err)
	}
}

func TestTransportFailuresPropagate(t *testing.T) {
	t.Parallel()

	// A swallowed 500 here reads as "this restaurant has no prices", which is
	// indistinguishable from a real empty result and wrong in a different way.
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("backend down"))
	})
	s := New(c)
	ctx := context.Background()
	from, to := day("2026-09-01"), day("2026-09-30")

	calls := map[string]func() error{
		"ListPrices":          func() error { _, _, e := s.ListPrices(ctx, from, to, nil, false, "", -1); return e },
		"ListMenuChanges":     func() error { _, _, e := s.ListMenuChanges(ctx, from, to, "", -1); return e },
		"MenuChangeByID":      func() error { _, e := s.MenuChangeByID(ctx, "doc-1"); return e },
		"MenuChangeByNumber":  func() error { _, e := s.MenuChangeByNumber(ctx, "MC-7"); return e },
		"SaveMenuChange":      func() error { _, e := s.SaveMenuChange(ctx, MenuChangeDocumentDto{}); return e },
		"ListPriceCategories": func() error { _, _, e := s.ListPriceCategories(ctx, false, nil, -1); return e },
		"PriceCategoryByID":   func() error { _, e := s.PriceCategoryByID(ctx, "pc-1"); return e },
		"ListPeriodSchedules": func() error { _, _, e := s.ListPeriodSchedules(ctx, false, nil, -1); return e },
		"PeriodScheduleByID":  func() error { _, e := s.PeriodScheduleByID(ctx, "s-1"); return e },
	}
	if len(calls) != 9 {
		t.Fatalf("the catalog lists 9 pricing endpoints, this table has %d", len(calls))
	}
	for name, call := range calls {
		if err := call(); err == nil {
			t.Errorf("%s swallowed a 500", name)
		}
	}
}
