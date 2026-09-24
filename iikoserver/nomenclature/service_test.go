package nomenclature

import (
	"context"
	"net/http"
	"net/url"
	"testing"

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

func TestFindProducts(t *testing.T) {
	t.Parallel()
	// An enveloped list, which is what v2 endpoints return.
	const body = `{"result":"SUCCESS","errors":[],"response":[
		{"id":"p1","name":"Пицца Маргарита","num":"0001","type":"DISH","mainUnit":"шт"},
		{"id":"p2","name":"Мука пшеничная","num":"0002","code":"MUKA","type":"GOODS","mainUnit":"кг"},
		{"id":"p3","name":"Тесто","num":"0003","type":"PREPARED","mainUnit":"кг"}]}`

	tests := []struct {
		name   string
		search string
		want   []string // ids, in order
	}{
		{name: "empty search returns everything", search: "", want: []string{"p1", "p2", "p3"}},
		{name: "matches on name, case-insensitively", search: "пицца", want: []string{"p1"}},
		{name: "matches on article", search: "0002", want: []string{"p2"}},
		{name: "matches on code", search: "muka", want: []string{"p2"}},
		{name: "substring anywhere in the name", search: "пшенич", want: []string{"p2"}},
		{name: "no match is empty, not an error", search: "суши", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) })
			got, err := s.FindProducts(context.Background(), tt.search, nil, false)
			if err != nil {
				t.Fatalf("FindProducts: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d products, want %d: %+v", len(got), len(tt.want), got)
			}
			for i, id := range tt.want {
				if got[i].ID != id {
					t.Errorf("result %d = %q, want %q", i, got[i].ID, id)
				}
			}
		})
	}
}

func TestFindProductsPassesTypeFiltersAndIncludeDeleted(t *testing.T) {
	t.Parallel()
	var q url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		_, _ = w.Write([]byte(`[]`))
	})
	if _, err := s.FindProducts(context.Background(), "", []string{TypeDish, TypeGoods}, true); err != nil {
		t.Fatal(err)
	}
	// Repeated key, not a comma-joined value.
	if got := q["types"]; len(got) != 2 || got[0] != "DISH" || got[1] != "GOODS" {
		t.Errorf("types = %v, want two repeated keys", got)
	}
	// includeDeleted defaults differ per endpoint upstream, so we always send it.
	if q.Get("includeDeleted") != "true" {
		t.Errorf("includeDeleted = %q, want explicit true", q.Get("includeDeleted"))
	}
}

func TestProductTypesExcludeUnsaveableKinds(t *testing.T) {
	t.Parallel()
	// OUTER and PETROL exist in the API but cannot be created, so offering them
	// in a tool enum would invite a guaranteed rejection.
	for _, v := range ProductTypes() {
		if v == TypeOuter || v == TypePetrol {
			t.Errorf("ProductTypes() must not offer %v", v)
		}
	}
	if len(ProductTypes()) == 0 {
		t.Fatal("ProductTypes() is empty")
	}
}

func TestTransportFailuresPropagate(t *testing.T) {
	t.Parallel()

	// A swallowed 500 here reads as "this restaurant sells nothing", which is
	// indistinguishable from a real empty catalogue and wrong in a worse way.
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("backend down"))
	})
	s := New(c)
	ctx := context.Background()
	ids := []string{"x-1"}

	calls := map[string]func() error{
		"FindProducts":          func() error { _, e := s.FindProducts(ctx, "", nil, false); return e },
		"ProductsXML":           func() error { _, e := s.ProductsXML(ctx, false, -1); return e },
		"SearchProductsXML":     func() error { _, e := s.SearchProductsXML(ctx, ProductSearchXML{}); return e },
		"ListProducts":          func() error { _, e := s.ListProducts(ctx, ProductFilter{}); return e },
		"QueryProducts":         func() error { _, e := s.QueryProducts(ctx, ProductFilter{}); return e },
		"SaveProduct":           func() error { _, e := s.SaveProduct(ctx, ProductDto{}, true, true); return e },
		"UpdateProduct":         func() error { _, e := s.UpdateProduct(ctx, ProductDto{ID: "p-1"}, false, false); return e },
		"DeleteProducts":        func() error { _, e := s.DeleteProducts(ctx, ids); return e },
		"RestoreProducts":       func() error { _, e := s.RestoreProducts(ctx, ids, false); return e },
		"ListGroups":            func() error { _, e := s.ListGroups(ctx, TreeFilter{}); return e },
		"QueryGroups":           func() error { _, e := s.QueryGroups(ctx, TreeFilter{}); return e },
		"SaveGroup":             func() error { _, e := s.SaveGroup(ctx, ProductGroupDto{}, true, true); return e },
		"UpdateGroup":           func() error { _, e := s.UpdateGroup(ctx, ProductGroupDto{ID: "g-1"}, false, false); return e },
		"DeleteProductsAndGrp":  func() error { _, e := s.DeleteProductsAndGroups(ctx, ids, nil); return e },
		"RestoreProductsAndGrp": func() error { _, e := s.RestoreProductsAndGroups(ctx, ids, nil, false); return e },
		"ListCategories":        func() error { _, e := s.ListCategories(ctx, TreeFilter{}); return e },
		"QueryCategories":       func() error { _, e := s.QueryCategories(ctx, TreeFilter{}); return e },
		"SaveCategory":          func() error { _, e := s.SaveCategory(ctx, "Салаты"); return e },
		"UpdateCategory":        func() error { _, e := s.UpdateCategory(ctx, "c-1", "Салаты"); return e },
		"DeleteCategory":        func() error { _, e := s.DeleteCategory(ctx, "c-1"); return e },
		"RestoreCategory":       func() error { _, e := s.RestoreCategory(ctx, "c-1"); return e },
		"ListProductScales":     func() error { _, e := s.ListProductScales(ctx, false, nil); return e },
		"QueryProductScales":    func() error { _, e := s.QueryProductScales(ctx, false, nil); return e },
		"ProductScaleByID":      func() error { _, e := s.ProductScaleByID(ctx, "sc-1"); return e },
		"SaveProductScale":      func() error { _, e := s.SaveProductScale(ctx, ProductScaleDto{}); return e },
		"UpdateProductScale":    func() error { _, e := s.UpdateProductScale(ctx, ProductScaleDto{ID: "sc-1"}); return e },
		"DeleteProductScales":   func() error { _, e := s.DeleteProductScales(ctx, ids); return e },
		"RestoreProductScales":  func() error { _, e := s.RestoreProductScales(ctx, ids); return e },
		"ProductScaleBinding":   func() error { _, e := s.ProductScaleBinding(ctx, "p-1"); return e },
		"BindProductScale":      func() error { _, e := s.BindProductScale(ctx, "p-1", ProductScaleBindingDto{ID: "sc-1"}); return e },
		"UnbindProductScale":    func() error { _, e := s.UnbindProductScale(ctx, "p-1"); return e },
		"ProductScalesFor":      func() error { _, e := s.ProductScalesFor(ctx, false, ids); return e },
		"QueryProductScalesFor": func() error { _, e := s.QueryProductScalesFor(ctx, false, ids); return e },
		"ListQuickMenus":        func() error { _, e := s.ListQuickMenus(ctx, QuickMenuFilter{}); return e },
		"QueryQuickMenus":       func() error { _, e := s.QueryQuickMenus(ctx, QuickMenuFilter{}); return e },
		"SaveQuickMenu":         func() error { _, e := s.SaveQuickMenu(ctx, QuickMenuDto{}); return e },
		"UpdateQuickMenu":       func() error { _, e := s.UpdateQuickMenu(ctx, QuickMenuDto{ID: "qm-1"}); return e },
		"DeleteQuickMenu":       func() error { _, e := s.DeleteQuickMenu(ctx, "qm-1"); return e },
		"LoadImage":             func() error { _, e := s.LoadImage(ctx, "img-1"); return e },
		"SaveImage":             func() error { _, e := s.SaveImage(ctx, "data"); return e },
		"DeleteImages":          func() error { _, e := s.DeleteImages(ctx, ids); return e },
	}
	for name, call := range calls {
		if err := call(); err == nil {
			t.Errorf("%s swallowed a 500", name)
		}
	}
}
