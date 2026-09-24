package nomenclature

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Nap20192/iiko-go/iikoserver/rest/resttest"
)

type seen struct {
	method      string
	url         url.URL
	body        string
	contentType string
}

func dial(t *testing.T, response string) (*Service, *seen) {
	t.Helper()
	got := &seen{}
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		*got = seen{method: r.Method, url: *r.URL, body: string(b), contentType: r.Header.Get("Content-Type")}
		_, _ = w.Write([]byte(response))
	})
	return New(c), got
}

func TestListProductsRepeatsFiltersAndKeepsTheNullParent(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `[{"id":"p-1","name":"Апельсин","type":"GOODS"}]`)
	prods, err := s.ListProducts(context.Background(), ProductFilter{
		IDs:       []string{"p-1", "p-2"},
		Types:     []string{TypeGoods, TypeDish},
		ParentIDs: []string{"g-1", ""},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(prods) != 1 || prods[0].Name != "Апельсин" {
		t.Fatalf("got %+v", prods)
	}
	q := got.url.Query()
	if got.method != http.MethodGet {
		t.Errorf("method %s", got.method)
	}
	// Upstream defaults flip per endpoint, so this is never left unsaid.
	if q.Get("includeDeleted") != "false" {
		t.Errorf("includeDeleted = %q, want an explicit false", q.Get("includeDeleted"))
	}
	if ids := q["ids"]; len(ids) != 2 || ids[1] != "p-2" {
		t.Errorf("ids = %v", ids)
	}
	if ts := q["types"]; len(ts) != 2 || ts[0] != "GOODS" {
		t.Errorf("types = %v", ts)
	}
	// "parentIds=" with an empty value is the documented filter for "no parent";
	// dropping the empty string silently turns it into "any parent".
	if ps := q["parentIds"]; len(ps) != 2 || ps[1] != "" {
		t.Errorf("parentIds = %v, want the empty null-filter value kept", ps)
	}
}

func TestListProductsRejectsWhatOnlyThePostTwinAccepts(t *testing.T) {
	t.Parallel()

	// codes and revisionFrom exist only on the POST form of this same path.
	// Sent on the GET they are ignored, and an incremental sync silently
	// re-reads the whole catalogue every poll.
	s, _ := dial(t, `[]`)
	rev := int64(42)
	for name, f := range map[string]ProductFilter{
		"codes":        {Codes: []string{"0001"}},
		"revisionFrom": {RevisionFrom: &rev},
	} {
		_, err := s.ListProducts(context.Background(), f)
		if err == nil || !strings.Contains(err.Error(), "QueryProducts") {
			t.Errorf("%s: got %v, want an error naming QueryProducts", name, err)
		}
	}
}

func TestQueryProductsPostsAFormNotJSON(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `[{"id":"p-1"}]`)
	rev := int64(42)
	if _, err := s.QueryProducts(context.Background(), ProductFilter{
		IncludeDeleted: true,
		Codes:          []string{"0001", "0002"},
		RevisionFrom:   &rev,
	}); err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodPost {
		t.Errorf("method %s", got.method)
	}
	// The POST twins are form-encoded, not JSON — PostV2 would send a body
	// iiko parses as an empty filter and answer with the whole catalogue.
	if !strings.HasPrefix(got.contentType, "application/x-www-form-urlencoded") {
		t.Errorf("Content-Type = %q", got.contentType)
	}
	form, err := url.ParseQuery(got.body)
	if err != nil {
		t.Fatal(err)
	}
	if form.Get("revisionFrom") != "42" || len(form["codes"]) != 2 || form.Get("includeDeleted") != "true" {
		t.Errorf("form was %q", got.body)
	}
}

func TestSaveProductSendsBothCodeSwitchesExplicitly(t *testing.T) {
	t.Parallel()

	// Both default to true upstream. A caller that wants to keep its own num
	// has to say so, and silence would mean iiko overwrites it.
	s, got := dial(t, `{"result":"SUCCESS","response":{"id":"p-1"}}`)
	p, err := s.SaveProduct(context.Background(),
		ProductDto{Name: "Апельсин", Num: "0001", MainUnit: "кг", Type: TypeGoods}, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "p-1" {
		t.Fatalf("got %+v", p)
	}
	q := got.url.Query()
	if q.Get("generateNomenclatureCode") != "false" || q.Get("generateFastCode") != "false" {
		t.Errorf("query was %s", got.url.RawQuery)
	}
}

func TestSaveProductEnforcesTheDocumentedFieldRules(t *testing.T) {
	t.Parallel()

	s, _ := dial(t, `{"result":"SUCCESS","response":{}}`)
	base := ProductDto{Name: "Апельсин", MainUnit: "кг", Type: TypeGoods}

	cases := []struct {
		name         string
		product      ProductDto
		generateNum  bool
		wantContains string
	}{
		{"num required when not generated", base, false, "num"},
		{"placeType required when in menu",
			func() ProductDto { p := base; p.Num = "1"; p.DefaultIncludeInMenu = true; return p }(),
			false, "placeType"},
		{"excludedSections must be empty when not in menu",
			func() ProductDto { p := base; p.ExcludedSections = []string{"s-1"}; return p }(),
			true, "excludedSections"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := s.SaveProduct(context.Background(), tc.product, tc.generateNum, true)
			if err == nil || !strings.Contains(err.Error(), tc.wantContains) {
				t.Fatalf("got %v, want an error about %s", err, tc.wantContains)
			}
		})
	}
}

func TestSaveProductSendsTheFieldTableSpellingOfDefaultIncludeInMenu(t *testing.T) {
	t.Parallel()

	// The docs spell it defaultIncludeInMenu in the field table and
	// defaultIncludedInMenu in the examples. We read both and write the table's.
	s, got := dial(t, `{"result":"SUCCESS","response":{"id":"p-1"}}`)
	if _, err := s.SaveProduct(context.Background(), ProductDto{
		Name: "Апельсин", MainUnit: "кг", Type: TypeGoods,
		DefaultIncludeInMenu: true, PlaceType: "place-1",
	}, true, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.body, `"defaultIncludeInMenu":true`) {
		t.Errorf("body was %s", got.body)
	}
	if strings.Contains(got.body, "defaultIncludedInMenu") {
		t.Errorf("the example spelling must not be written: %s", got.body)
	}
}

func TestSaveProductTreatsResultErrorAsFailure(t *testing.T) {
	t.Parallel()

	// HTTP 200 with result=ERROR is how save rejects a product.
	s, _ := dial(t, `{"result":"ERROR","errors":[{"code":"COOKING_PLACE_EMPTY_FOR_SALE_DISH","value":"не задано место приготовления"}]}`)
	_, err := s.SaveProduct(context.Background(),
		ProductDto{Name: "Салат", MainUnit: "порц", Type: TypeDish}, true, true)
	if err == nil || !strings.Contains(err.Error(), "COOKING_PLACE_EMPTY_FOR_SALE_DISH") {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateProductSendsTheOverrideSwitches(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `{"result":"SUCCESS","response":{"id":"p-1"}}`)
	if _, err := s.UpdateProduct(context.Background(),
		ProductDto{ID: "p-1", Name: "Апельсин", MainUnit: "кг", Type: TypeGoods}, true, false); err != nil {
		t.Fatal(err)
	}
	q := got.url.Query()
	if q.Get("overrideFastCode") != "true" || q.Get("overrideNomenclatureCode") != "false" {
		t.Errorf("query was %s", got.url.RawQuery)
	}
	// update addresses an existing element, so the id has to survive marshalling.
	if !strings.Contains(got.body, `"id":"p-1"`) {
		t.Errorf("body was %s", got.body)
	}
	if _, err := s.UpdateProduct(context.Background(), ProductDto{Name: "x", MainUnit: "кг", Type: TypeGoods}, false, false); err == nil {
		t.Error("id is required on update")
	}
}

func TestDeleteAndRestoreProductsUseTheIDListBody(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `{"result":"SUCCESS","response":[{"id":"p-1","deleted":true}]}`)
	prods, err := s.DeleteProducts(context.Background(), []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(prods) != 1 || !prods[0].Deleted {
		t.Fatalf("got %+v", prods)
	}
	if !strings.Contains(got.body, `"items":[{"id":"p-1"}]`) {
		t.Errorf("body was %s", got.body)
	}

	if _, err := s.RestoreProducts(context.Background(), []string{"p-1"}, true); err != nil {
		t.Fatal(err)
	}
	if got.url.Query().Get("overrideNomenclatureCode") != "true" {
		t.Errorf("query was %s", got.url.RawQuery)
	}

	for name, call := range map[string]func() error{
		"delete":  func() error { _, e := s.DeleteProducts(context.Background(), nil); return e },
		"restore": func() error { _, e := s.RestoreProducts(context.Background(), nil, false); return e },
	} {
		if err := call(); err == nil {
			t.Errorf("%s: an empty id list would touch nothing and read as success", name)
		}
	}
}

func TestSoftDeleteConflictsComeBackVerbatim(t *testing.T) {
	t.Parallel()

	// Deleting a deleted product, or a group whose children stay, is a 409 with
	// a plain-text body naming the offending ids — not a v2 errors[] entry.
	// Re-wording it locally would lose the ids, which are the actionable part.
	const body = "Could not delete already deleted products: [p-1, p-2]"
	c := resttest.NewFake(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(body))
	})
	_, err := New(c).DeleteProducts(context.Background(), []string{"p-1"})
	if err == nil || !strings.Contains(err.Error(), body) {
		t.Fatalf("got %v, want the server's own text", err)
	}
}

func TestProductsXMLReadsTheLegacyRoot(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `<productDtoes><productDto><id>p-1</id><name>Апельсин</name>
		<productType>GOODS</productType></productDto></productDtoes>`)
	prods, err := s.ProductsXML(context.Background(), true, 17)
	if err != nil {
		t.Fatal(err)
	}
	if len(prods) != 1 || prods[0].Name != "Апельсин" || prods[0].ProductType != TypeGoods {
		t.Fatalf("got %+v", prods)
	}
	q := got.url.Query()
	if q.Get("includeDeleted") != "true" || q.Get("revisionFrom") != "17" {
		t.Errorf("query was %s", got.url.RawQuery)
	}
}

func TestSearchProductsXMLSendsRegexFilters(t *testing.T) {
	t.Parallel()

	// Every filter here is a regular expression, not a literal — the docs say so
	// for all seven, and a caller passing a bare name gets substring behaviour.
	s, got := dial(t, `<productDtoes/>`)
	if _, err := s.SearchProductsXML(context.Background(), ProductSearchXML{
		IncludeDeleted: true,
		Name:           "^Апельсин",
		ProductType:    "GOODS|DISH",
	}); err != nil {
		t.Fatal(err)
	}
	q := got.url.Query()
	if q.Get("name") != "^Апельсин" || q.Get("productType") != "GOODS|DISH" {
		t.Errorf("query was %s", got.url.RawQuery)
	}
	// Absent filters must not be sent as empty regexes, which match everything.
	for _, k := range []string{"code", "mainUnit", "num", "cookingPlaceType", "productGroupType"} {
		if _, ok := q[k]; ok {
			t.Errorf("%s was sent empty", k)
		}
	}
}
