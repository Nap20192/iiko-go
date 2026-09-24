package nomenclature

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestListGroupsAndCategoriesAlwaysSendIncludeDeleted(t *testing.T) {
	t.Parallel()

	t.Run("groups", func(t *testing.T) {
		t.Parallel()
		s, got := dial(t, `[{"id":"g-1","name":"Фрукты"}]`)
		groups, err := s.ListGroups(context.Background(), TreeFilter{
			IDs: []string{"g-1"}, ParentIDs: []string{""}, Codes: []string{"0001"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(groups) != 1 || groups[0].Name != "Фрукты" {
			t.Fatalf("got %+v", groups)
		}
		q := got.url.Query()
		if q.Get("includeDeleted") != "false" || q.Get("revisionFrom") != "-1" {
			t.Errorf("query was %s", got.url.RawQuery)
		}
		// Unlike products, the group list's GET accepts codes too.
		if q.Get("codes") != "0001" {
			t.Errorf("codes = %q", q.Get("codes"))
		}
		if ps := q["parentIds"]; len(ps) != 1 || ps[0] != "" {
			t.Errorf("parentIds = %v, want the null-filter value kept", ps)
		}
	})

	t.Run("categories", func(t *testing.T) {
		t.Parallel()
		s, got := dial(t, `[{"id":"c-1","name":"Салаты","rootType":"ProductCategory"}]`)
		cats, err := s.ListCategories(context.Background(), TreeFilter{IncludeDeleted: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(cats) != 1 || cats[0].Name != "Салаты" {
			t.Fatalf("got %+v", cats)
		}
		if got.url.Query().Get("includeDeleted") != "true" {
			t.Errorf("query was %s", got.url.RawQuery)
		}
	})
}

func TestQueryGroupsAndCategoriesPostForms(t *testing.T) {
	t.Parallel()

	rev := int64(9)
	t.Run("groups", func(t *testing.T) {
		t.Parallel()
		s, got := dial(t, `[]`)
		if _, err := s.QueryGroups(context.Background(), TreeFilter{RevisionFrom: &rev}); err != nil {
			t.Fatal(err)
		}
		if got.method != http.MethodPost || !strings.HasPrefix(got.contentType, "application/x-www-form-urlencoded") {
			t.Errorf("%s %s", got.method, got.contentType)
		}
		form, _ := url.ParseQuery(got.body)
		if form.Get("revisionFrom") != "9" {
			t.Errorf("form was %q", got.body)
		}
	})
	t.Run("categories", func(t *testing.T) {
		t.Parallel()
		s, got := dial(t, `[]`)
		if _, err := s.QueryCategories(context.Background(), TreeFilter{IDs: []string{"c-1"}}); err != nil {
			t.Fatal(err)
		}
		form, _ := url.ParseQuery(got.body)
		if form.Get("ids") != "c-1" {
			t.Errorf("form was %q", got.body)
		}
	})
}

func TestUpdateGroupCarriesTheID(t *testing.T) {
	t.Parallel()

	// update addresses the group by id, so it has to survive marshalling, while
	// save must not carry one — the server assigns it.
	s, got := dial(t, `{"result":"SUCCESS","response":{"id":"g-1","name":"Фрукты"}}`)
	g, err := s.UpdateGroup(context.Background(),
		ProductGroupDto{ID: "g-1", Name: "Фрукты"}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "Фрукты" {
		t.Fatalf("got %+v", g)
	}
	if !strings.Contains(got.body, `"id":"g-1"`) || !strings.Contains(got.body, `"name":"Фрукты"`) {
		t.Errorf("body was %s", got.body)
	}
	q := got.url.Query()
	if q.Get("overrideFastCode") != "true" || q.Get("overrideNomenclatureCode") != "false" {
		t.Errorf("query was %s", got.url.RawQuery)
	}
	if _, err := s.UpdateGroup(context.Background(), ProductGroupDto{Name: "x"}, false, false); err == nil {
		t.Error("id is required on update")
	}
}

func TestSaveGroupSendsTheCodeSwitches(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `{"result":"SUCCESS","response":{"id":"g-1"}}`)
	if _, err := s.SaveGroup(context.Background(), ProductGroupDto{Name: "Фрукты"}, false, true); err != nil {
		t.Fatal(err)
	}
	q := got.url.Query()
	if q.Get("generateNomenclatureCode") != "false" || q.Get("generateFastCode") != "true" {
		t.Errorf("query was %s", got.url.RawQuery)
	}
	if strings.Contains(got.body, `"id"`) {
		t.Errorf("save must not carry an id: %s", got.body)
	}
}

func TestDeleteGroupTakesProductsAndGroupsTogether(t *testing.T) {
	t.Parallel()

	// One call deletes both, because iiko refuses to delete a group whose
	// children stay behind — the two lists have to travel together.
	s, got := dial(t, `{"result":"SUCCESS","response":{
		"products":[{"id":"p-1","deleted":true}],
		"productGroups":[{"id":"g-1","deleted":true}]}}`)
	res, err := s.DeleteProductsAndGroups(context.Background(), []string{"p-1"}, []string{"g-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Products) != 1 || len(res.Groups) != 1 || !res.Groups[0].Deleted {
		t.Fatalf("got %+v", res)
	}
	for _, want := range []string{`"products":{"items":[{"id":"p-1"}]}`, `"productGroups":{"items":[{"id":"g-1"}]}`} {
		if !strings.Contains(got.body, want) {
			t.Errorf("missing %s in %s", want, got.body)
		}
	}
	if _, err := s.DeleteProductsAndGroups(context.Background(), nil, nil); err == nil {
		t.Error("both lists empty would touch nothing and read as success")
	}
}

func TestRestoreGroupSendsOverrideAndAcceptsOneListOnly(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `{"result":"SUCCESS","response":{"productGroups":[{"id":"g-1"}]}}`)
	res, err := s.RestoreProductsAndGroups(context.Background(), nil, []string{"g-1"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Groups) != 1 || len(res.Products) != 0 {
		t.Fatalf("got %+v", res)
	}
	if got.url.Query().Get("overrideNomenclatureCode") != "true" {
		t.Errorf("query was %s", got.url.RawQuery)
	}
}

func TestCategoryWritesUseTheirOwnBareBodies(t *testing.T) {
	t.Parallel()

	// Categories do not use the {items:[{id}]} list every other delete here
	// takes: save is {name}, update is {id,name}, delete and restore are {id}.
	cases := []struct {
		name string
		call func(*Service) error
		want string
	}{
		{"save", func(s *Service) error { _, e := s.SaveCategory(context.Background(), "Салаты"); return e },
			`{"name":"Салаты"}`},
		{"update", func(s *Service) error {
			_, e := s.UpdateCategory(context.Background(), "c-1", "Салаты")
			return e
		},
			`{"id":"c-1","name":"Салаты"}`},
		{"delete", func(s *Service) error { _, e := s.DeleteCategory(context.Background(), "c-1"); return e },
			`{"id":"c-1"}`},
		{"restore", func(s *Service) error { _, e := s.RestoreCategory(context.Background(), "c-1"); return e },
			`{"id":"c-1"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, got := dial(t, `{"result":"SUCCESS","response":{"id":"c-1","name":"Салаты"}}`)
			if err := tc.call(s); err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(got.body) != tc.want {
				t.Errorf("body = %s, want %s", got.body, tc.want)
			}
		})
	}
}

func TestCategoryWritesRejectEmptyArguments(t *testing.T) {
	t.Parallel()

	// The server answers a blank name with plain text, not a v2 errors[] entry,
	// so it is cheaper to refuse here than to decode that.
	s, _ := dial(t, `{"result":"SUCCESS","response":{}}`)
	for name, call := range map[string]func() error{
		"save blank name":   func() error { _, e := s.SaveCategory(context.Background(), "   "); return e },
		"update blank name": func() error { _, e := s.UpdateCategory(context.Background(), "c-1", ""); return e },
		"update no id":      func() error { _, e := s.UpdateCategory(context.Background(), "", "Салаты"); return e },
		"delete no id":      func() error { _, e := s.DeleteCategory(context.Background(), ""); return e },
		"restore no id":     func() error { _, e := s.RestoreCategory(context.Background(), ""); return e },
	} {
		if err := call(); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
