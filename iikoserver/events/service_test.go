package events

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
	"github.com/Nap20192/iiko-go/iikoserver/rest/resttest"
)

func newFake(t *testing.T, body string) (*Service, *url.URL, *string) {
	t.Helper()
	got, sent := &url.URL{}, new(string)
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		*got = *r.URL
		b, _ := io.ReadAll(r.Body)
		*sent = string(b)
		_, _ = w.Write([]byte(body))
	})
	return New(c), got, sent
}

func stamp(s string) time.Time {
	t, err := time.Parse(rest.OlapMs, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestListEventsSendsMillisecondStamps(t *testing.T) {
	t.Parallel()

	// Events are the only /resto/api family outside OLAP that wants
	// yyyy-MM-ddTHH:mm:ss.SSS; sending the v2 yyyy-MM-dd form returns nothing.
	s, got, _ := newFake(t, `<eventsList><revision>10</revision></eventsList>`)
	if _, err := s.ListEvents(context.Background(),
		stamp("2026-09-01T00:00:00.000"), stamp("2026-09-02T00:00:00.000")); err != nil {
		t.Fatal(err)
	}
	q := got.Query()
	if q.Get("from_time") != "2026-09-01T00:00:00.000" {
		t.Errorf("from_time = %q", q.Get("from_time"))
	}
	// to_time is exclusive upstream and is sent verbatim: nudging it by a
	// millisecond to "make it inclusive" would double-count a boundary event.
	if q.Get("to_time") != "2026-09-02T00:00:00.000" {
		t.Errorf("to_time = %q", q.Get("to_time"))
	}
}

func TestListEventsSinceNeverSendsATimeBound(t *testing.T) {
	t.Parallel()

	// The docs are explicit: to_time must not accompany from_rev. Sending both
	// silently drops the revision cursor and re-reads the window every poll.
	s, got, _ := newFake(t, `<eventsList><revision>10</revision></eventsList>`)
	if _, err := s.ListEventsSince(context.Background(), 11); err != nil {
		t.Fatal(err)
	}
	q := got.Query()
	if q.Get("from_rev") != "11" {
		t.Errorf("from_rev = %q", q.Get("from_rev"))
	}
	if _, ok := q["to_time"]; ok {
		t.Error("to_time must not be sent in revision mode")
	}
	if _, ok := q["from_time"]; ok {
		t.Error("from_time must not be sent in revision mode")
	}
}

func TestNextRevisionIsOneMoreThanTheResponse(t *testing.T) {
	t.Parallel()

	// <revision> means "returned up to and including this one", so polling with
	// the same number redelivers the last batch and +2 skips a batch.
	if got := NextRevision(&EventsListDto{Revision: 10}); got != 11 {
		t.Fatalf("got %d, want 11", got)
	}
}

func TestDedupKeepsFirstByEventID(t *testing.T) {
	t.Parallel()

	// Redelivery under a new revision is possible but not documented as
	// impossible; the event UUID is the documented dedup key.
	in := []EventDto{{ID: "a", Type: "orderPaid"}, {ID: "b"}, {ID: "a", Type: "duplicate"}}
	out := Dedup(in)
	if len(out) != 2 || out[0].Type != "orderPaid" || out[1].ID != "b" {
		t.Fatalf("got %+v", out)
	}
}

func TestFilterEventsSendsTheDocumentedRequestShape(t *testing.T) {
	t.Parallel()

	s, _, sent := newFake(t, `<eventsList/>`)
	if _, err := s.FilterEvents(context.Background(),
		[]string{"orderPaid", "orderCancelPrecheque"}, []string{"175658"}); err != nil {
		t.Fatal(err)
	}
	// <events><event>… is a nested list, not repeated <events> elements; the
	// catalog's flat spelling would produce a body iiko ignores.
	for _, want := range []string{
		"<eventsRequestData>",
		"<events><event>orderPaid</event><event>orderCancelPrecheque</event></events>",
		"<orderNums><orderNum>175658</orderNum></orderNums>",
	} {
		if !strings.Contains(*sent, want) {
			t.Errorf("missing %s in %s", want, *sent)
		}
	}
}

func TestAddEventsNeverSendsAnID(t *testing.T) {
	t.Parallel()

	// The server assigns the UUID and returns it; a submitted id is rejected.
	s, got, sent := newFake(t, `<eventsList><event><id>srv-1</id></event></eventsList>`)
	out, err := s.AddEvents(context.Background(), []EventDto{
		{ID: "client-made-this-up", Type: "banana", Date: "2026-09-01T12:00:00.000"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(*sent, "client-made-this-up") {
		t.Errorf("id was sent: %s", *sent)
	}
	if !strings.Contains(*sent, "<eventsList>") || !strings.Contains(*sent, "<type>banana</type>") {
		t.Errorf("body was %s", *sent)
	}
	if got.Path != "/resto"+EndpointEventsAdd {
		t.Errorf("path %s", got.Path)
	}
	if len(out.Event) != 1 || out.Event[0].ID != "srv-1" {
		t.Errorf("server id must come back: %+v", out)
	}
}

func TestMissingModuleOrRightIsNamedInTheError(t *testing.T) {
	t.Parallel()

	// 403 here almost always means module 2200 or B_VTJ, not a bad password,
	// and the two are fixed in different places by different people.
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Access denied"))
	})
	_, err := New(c).Metadata(context.Background())
	if err == nil {
		t.Fatal("want an error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "B_VTJ") || !strings.Contains(msg, "2200") {
		t.Fatalf("error must name the right and the module, got %q", msg)
	}
}

func TestMetadataDecodesTheEventTree(t *testing.T) {
	t.Parallel()

	s, got, _ := newFake(t, `<groupsList><group><id>g1</id><name>Заказы</name>
		<type><id>orderPaid</id><name>Заказ оплачен</name><severity>1</severity></type></group></groupsList>`)
	tree, err := s.Metadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Group) != 1 || len(tree.Group[0].Type) != 1 || tree.Group[0].Type[0].ID != "orderPaid" {
		t.Fatalf("got %+v", tree)
	}
	if got.Path != "/resto"+EndpointEventsMetadata {
		t.Errorf("path %s", got.Path)
	}
}

func TestFilterMetadataPostsOnlyTheTypeList(t *testing.T) {
	t.Parallel()

	s, _, sent := newFake(t, `<groupsList/>`)
	if _, err := s.FilterMetadata(context.Background(), []string{"orderPaid"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(*sent, "<events><event>orderPaid</event></events>") {
		t.Errorf("body was %s", *sent)
	}
	if strings.Contains(*sent, "orderNums") {
		t.Error("the metadata filter takes no order numbers")
	}
}

func TestListSessionsReturnsRawXML(t *testing.T) {
	t.Parallel()

	// The docs describe the shift payload in prose only — no field table, no
	// example — so naming fields here would be invention.
	const body = `<sessions><session><number>7</number></session></sessions>`
	s, got, _ := newFake(t, body)
	raw, err := s.ListSessions(context.Background(), stamp("2026-09-01T00:00:00.000"), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != body {
		t.Fatalf("got %q", raw)
	}
	if _, ok := got.Query()["to_time"]; ok {
		t.Error("a zero upper bound must be omitted, not sent as a zero date")
	}
}
