package nomenclature

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestListQuickMenusUsesTheSingularIDParameter(t *testing.T) {
	t.Parallel()

	// Every sibling list in this domain spells the filter "ids". Quick menus
	// spell it "id", and the plural is silently ignored — you get every menu.
	s, got := dial(t, `[{"id":"qm-1","departmentId":"d-1"}]`)
	menus, err := s.ListQuickMenus(context.Background(), QuickMenuFilter{
		IDs:           []string{"qm-1"},
		DepartmentIDs: []string{"d-1"},
		SectionIDs:    []string{SectionDepartmentWide},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(menus) != 1 || menus[0].DepartmentID != "d-1" {
		t.Fatalf("got %+v", menus)
	}
	q := got.url.Query()
	if _, plural := q["ids"]; plural {
		t.Error("the parameter is id, not ids")
	}
	if q.Get("id") != "qm-1" {
		t.Errorf("id = %q", q.Get("id"))
	}
	// sectionId=null is the documented filter for the department-wide menu.
	if q.Get("sectionId") != "null" {
		t.Errorf("sectionId = %q", q.Get("sectionId"))
	}
	if q.Get("includeDeleted") != "false" {
		t.Errorf("includeDeleted = %q", q.Get("includeDeleted"))
	}
	// Unlike its siblings the docs state no default here, so nothing is invented.
	if _, ok := q["revisionFrom"]; ok {
		t.Error("revisionFrom has no documented default and must not be sent unasked")
	}
}

func TestQueryQuickMenusPostsAForm(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `[]`)
	rev := int64(4)
	if _, err := s.QueryQuickMenus(context.Background(), QuickMenuFilter{RevisionFrom: &rev}); err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodPost || !strings.HasPrefix(got.contentType, "application/x-www-form-urlencoded") {
		t.Errorf("%s %s", got.method, got.contentType)
	}
	if !strings.Contains(got.body, "revisionFrom=4") {
		t.Errorf("form was %q", got.body)
	}
}

func TestSaveQuickMenuKeepsZeroCoordinates(t *testing.T) {
	t.Parallel()

	// page, x and y are all 0-based and 0 is the first page, column and row.
	// The generated marshaller carries omitempty on each, so a label in the
	// top-left corner would arrive with no coordinates at all and land wherever
	// the server defaults — the whole menu shifts, silently.
	s, got := dial(t, `{"id":"qm-1"}`)
	if _, err := s.SaveQuickMenu(context.Background(), QuickMenuDto{
		DepartmentID: "d-1",
		Labels:       []QuickLabelDto{{Page: 0, X: 0, Y: 0, EntityID: "p-1"}},
	}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"page":0`, `"x":0`, `"y":0`, `"entityId":"p-1"`} {
		if !strings.Contains(got.body, want) {
			t.Errorf("missing %s in %s", want, got.body)
		}
	}
	// entityType is response-only: the server resolves it from entityId.
	if strings.Contains(got.body, "entityType") {
		t.Errorf("entityType must not be sent: %s", got.body)
	}
}

func TestQuickMenuDayFollowsDependsOnWeekDay(t *testing.T) {
	t.Parallel()

	// day is null on every label of a menu that is not day-of-week specific,
	// and 0 means Monday when it is. Sending 0 for a non-specific menu would
	// pin the whole menu to Mondays.
	t.Run("not day specific", func(t *testing.T) {
		t.Parallel()
		s, got := dial(t, `{"id":"qm-1"}`)
		if _, err := s.SaveQuickMenu(context.Background(), QuickMenuDto{
			DepartmentID: "d-1", DependsOnWeekDay: false,
			Labels: []QuickLabelDto{{Day: 0, EntityID: "p-1"}},
		}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got.body, `"day":null`) {
			t.Errorf("body was %s", got.body)
		}
	})
	t.Run("day specific", func(t *testing.T) {
		t.Parallel()
		s, got := dial(t, `{"id":"qm-1"}`)
		if _, err := s.SaveQuickMenu(context.Background(), QuickMenuDto{
			DepartmentID: "d-1", DependsOnWeekDay: true,
			Labels: []QuickLabelDto{{Day: 0, EntityID: "p-1"}},
		}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got.body, `"day":0`) {
			t.Errorf("body was %s", got.body)
		}
	})
}

func TestQuickLabelDayIsMondayZero(t *testing.T) {
	t.Parallel()

	// 0=Monday..6=Sunday here, while periodSchedules in pkg/iikoserver/pricing
	// numbers 1=Monday..7=Sunday. Same repo, same API, two conventions.
	for n, want := range map[int]time.Weekday{0: time.Monday, 5: time.Saturday, 6: time.Sunday} {
		got, err := QuickLabelDay(QuickLabelDto{Day: n})
		if err != nil || got != want {
			t.Errorf("%d -> %v (%v), want %v", n, got, err, want)
		}
	}
	if _, err := QuickLabelDay(QuickLabelDto{Day: 7}); err == nil {
		t.Error("7 is Sunday under the periodSchedules numbering and out of range here")
	}
	if _, err := QuickLabelDay(QuickLabelDto{Day: -1}); err == nil {
		t.Error("-1 must be rejected")
	}
}

func TestQuickMenuRejectsOutOfGridLabels(t *testing.T) {
	t.Parallel()

	s, _ := dial(t, `{"id":"qm-1"}`)
	cases := map[string]QuickLabelDto{
		"page": {Page: 3, EntityID: "p-1"},
		"x":    {X: 3, EntityID: "p-1"},
		"y":    {Y: 8, EntityID: "p-1"},
	}
	for name, label := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := s.SaveQuickMenu(context.Background(), QuickMenuDto{
				DepartmentID: "d-1", Labels: []QuickLabelDto{label},
			})
			if err == nil || !strings.Contains(err.Error(), name) {
				t.Fatalf("got %v, want an error naming %s", err, name)
			}
		})
	}
}

func TestUpdateAndDeleteQuickMenuCarryTheID(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `{"id":"qm-1"}`)
	if _, err := s.UpdateQuickMenu(context.Background(), QuickMenuDto{
		ID: "qm-1", DepartmentID: "d-1",
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.body, `"id":"qm-1"`) {
		t.Errorf("body was %s", got.body)
	}
	if _, err := s.UpdateQuickMenu(context.Background(), QuickMenuDto{DepartmentID: "d-1"}); err == nil {
		t.Error("id is required on update")
	}

	if _, err := s.DeleteQuickMenu(context.Background(), "qm-1"); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(got.body) != `{"id":"qm-1"}` {
		t.Errorf("delete body = %s, want a bare id", got.body)
	}
	if _, err := s.DeleteQuickMenu(context.Background(), ""); err == nil {
		t.Error("id is required")
	}
}

func TestImagesRoundTripBase64(t *testing.T) {
	t.Parallel()

	const payload = "/9j/4AAQSkZJRgABAQAAAQABAAD/2w=="

	t.Run("load", func(t *testing.T) {
		t.Parallel()
		s, got := dial(t, `{"id":"img-1","data":"`+payload+`"}`)
		img, err := s.LoadImage(context.Background(), "img-1")
		if err != nil {
			t.Fatal(err)
		}
		if img.Data != payload {
			t.Fatalf("got %+v", img)
		}
		if got.url.Query().Get("imageId") != "img-1" {
			t.Errorf("query was %s", got.url.RawQuery)
		}
		if _, err := s.LoadImage(context.Background(), ""); err == nil {
			t.Error("imageId is required")
		}
	})

	t.Run("save", func(t *testing.T) {
		t.Parallel()
		s, got := dial(t, `{"result":"SUCCESS","response":{"id":"img-1","data":"`+payload+`"}}`)
		img, err := s.SaveImage(context.Background(), payload)
		if err != nil {
			t.Fatal(err)
		}
		if img.ID != "img-1" {
			t.Fatalf("got %+v", img)
		}
		// The request field table names the Base64 payload "id", not "data" —
		// the response's id is the real UUID. Sending "data" writes nothing.
		if !strings.Contains(got.body, `"id":"`+payload+`"`) {
			t.Errorf("body was %s", got.body)
		}
		if _, err := s.SaveImage(context.Background(), " "); err == nil {
			t.Error("an empty payload is not an image")
		}
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		s, got := dial(t, `{"result":"SUCCESS","response":{"items":[{"id":"img-1"}]}}`)
		ids, err := s.DeleteImages(context.Background(), []string{"img-1"})
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) != 1 || ids[0] != "img-1" {
			t.Fatalf("got %v", ids)
		}
		if !strings.Contains(got.body, `"items":[{"id":"img-1"}]`) {
			t.Errorf("body was %s", got.body)
		}
		if _, err := s.DeleteImages(context.Background(), nil); err == nil {
			t.Error("an empty id list would delete nothing")
		}
	})
}
