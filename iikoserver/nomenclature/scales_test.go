package nomenclature

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// TestScaleBindingVerbs is the one place three operations share a URL. The docs
// never say they are different calls — only the page's CSS-tagged request blocks
// reveal it — so the verb is the whole distinction and nothing else checks it.
func got2Query(s *seen) url.Values { return s.url.Query() }

func TestScaleBindingVerbs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		want  string
		reply string
		call  func(*Service) error
	}{
		{"read", http.MethodGet, `{"result":"SUCCESS","response":{"id":"sc-1","name":"Размеры"}}`,
			func(s *Service) error { _, e := s.ProductScaleBinding(context.Background(), "p-1"); return e }},
		{"bind", http.MethodPost, `{"id":"sc-1","name":"Размеры"}`,
			func(s *Service) error {
				_, e := s.BindProductScale(context.Background(), "p-1", ProductScaleBindingDto{ID: "sc-1"})
				return e
			}},
		{"unbind", http.MethodDelete, `sc-1`,
			func(s *Service) error { _, e := s.UnbindProductScale(context.Background(), "p-1"); return e }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, got := dial(t, tc.reply)
			if err := tc.call(s); err != nil {
				t.Fatal(err)
			}
			if got.method != tc.want {
				t.Errorf("used %s, catalog says %s", got.method, tc.want)
			}
			if got.url.Path != "/resto/api/v2/entities/products/p-1/productScale" {
				t.Errorf("path %s", got.url.Path)
			}
		})
	}
}

func TestProductScaleBindingIsNilWhenNothingIsBound(t *testing.T) {
	t.Parallel()

	// An unbound product answers with response:null inside a SUCCESS envelope.
	// Reading that as an empty scale would claim the product has zero sizes.
	s, _ := dial(t, `{"result":"SUCCESS","errors":null,"response":null}`)
	scale, err := s.ProductScaleBinding(context.Background(), "p-1")
	if err != nil {
		t.Fatal(err)
	}
	if scale != nil {
		t.Fatalf("got %+v, want nil", scale)
	}
}

func TestUnbindReturnsTheScaleID(t *testing.T) {
	t.Parallel()

	// DELETE unbinds the scale from the product; it does not delete the scale
	// definition, which is what DeleteProductScales does.
	s, _ := dial(t, "  sc-1\n")
	id, err := s.UnbindProductScale(context.Background(), "p-1")
	if err != nil {
		t.Fatal(err)
	}
	if id != "sc-1" {
		t.Fatalf("got %q", id)
	}
}

func TestBindingSendsOnlyTheBindingFields(t *testing.T) {
	t.Parallel()

	// The binding carries the scale id plus per-size disabled and factors. Name,
	// shortName, priority and default belong to the definition, not the binding.
	s, got := dial(t, `{"id":"sc-1"}`)
	if _, err := s.BindProductScale(context.Background(), "p-1", ProductScaleBindingDto{
		ID:           "sc-1",
		ProductSizes: []ProductSizeDto{{ID: "sz-1", Disabled: true}},
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.body, `"id":"sc-1"`) || !strings.Contains(got.body, `"disabled":true`) {
		t.Errorf("body was %s", got.body)
	}
	if _, err := s.BindProductScale(context.Background(), "", ProductScaleBindingDto{ID: "sc-1"}); err == nil {
		t.Error("productId is required")
	}
	if _, err := s.BindProductScale(context.Background(), "p-1", ProductScaleBindingDto{}); err == nil {
		t.Error("the scale id is the point of the binding and is required")
	}
}

func TestListProductScalesRepeatsIDs(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `[{"id":"sc-1","name":"Размеры"}]`)
	scales, err := s.ListProductScales(context.Background(), false, []string{"sc-1", "sc-2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(scales) != 1 || scales[0].Name != "Размеры" {
		t.Fatalf("got %+v", scales)
	}
	q := got.url.Query()
	if q.Get("includeDeleted") != "false" {
		t.Errorf("includeDeleted = %q", q.Get("includeDeleted"))
	}
	if ids := q["ids"]; len(ids) != 2 {
		t.Errorf("ids = %v", ids)
	}
}

func TestQueryProductScalesPostsAForm(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `[]`)
	if _, err := s.QueryProductScales(context.Background(), true, []string{"sc-1"}); err != nil {
		t.Fatal(err)
	}
	if got.method != http.MethodPost || !strings.HasPrefix(got.contentType, "application/x-www-form-urlencoded") {
		t.Errorf("%s %s", got.method, got.contentType)
	}
	if !strings.Contains(got.body, "ids=sc-1") || !strings.Contains(got.body, "includeDeleted=true") {
		t.Errorf("form was %q", got.body)
	}
}

func TestProductScaleByIDReadsABareObject(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `{"id":"sc-1","name":"Размеры","productSizes":[{"id":"sz-1","name":"S"}]}`)
	scale, err := s.ProductScaleByID(context.Background(), "sc-1")
	if err != nil {
		t.Fatal(err)
	}
	if scale.Name != "Размеры" || len(scale.ProductSizes) != 1 {
		t.Fatalf("got %+v", scale)
	}
	if got.url.Path != "/resto/api/v2/entities/productScales/sc-1" {
		t.Errorf("path %s", got.url.Path)
	}
	if _, err := s.ProductScaleByID(context.Background(), ""); err == nil {
		t.Error("id is required")
	}
}

func TestUpdateProductScaleKeepsSizeIDs(t *testing.T) {
	t.Parallel()

	// A productSizes entry without an id creates a new size. Dropping the id on
	// the way out would silently duplicate every size on each update.
	s, got := dial(t, `{"result":"SUCCESS","response":{"id":"sc-1"}}`)
	if _, err := s.UpdateProductScale(context.Background(), ProductScaleDto{
		ID: "sc-1", Name: "Размеры",
		ProductSizes: []ProductSizeDto{{ID: "sz-1", Name: "S"}},
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.body, `"id":"sz-1"`) || !strings.Contains(got.body, `"id":"sc-1"`) {
		t.Errorf("body was %s", got.body)
	}
	if _, err := s.UpdateProductScale(context.Background(), ProductScaleDto{Name: "x"}); err == nil {
		t.Error("id is required on update")
	}
}

func TestDeleteAndRestoreProductScalesUseTheIDList(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `{"result":"SUCCESS","response":[{"id":"sc-1","deleted":true}]}`)
	scales, err := s.DeleteProductScales(context.Background(), []string{"sc-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(scales) != 1 || !scales[0].Deleted {
		t.Fatalf("got %+v", scales)
	}
	if !strings.Contains(got.body, `"items":[{"id":"sc-1"}]`) {
		t.Errorf("body was %s", got.body)
	}
	if _, err := s.RestoreProductScales(context.Background(), []string{"sc-1"}); err != nil {
		t.Fatal(err)
	}
	for name, call := range map[string]func() error{
		"delete":  func() error { _, e := s.DeleteProductScales(context.Background(), nil); return e },
		"restore": func() error { _, e := s.RestoreProductScales(context.Background(), nil); return e },
	} {
		if err := call(); err == nil {
			t.Errorf("%s: an empty id list would touch nothing", name)
		}
	}
}

func TestProductScalesForDecodesAMapWithNulls(t *testing.T) {
	t.Parallel()

	// The batch answers with an object keyed by productId, not an array, and a
	// product with no scale bound maps to null rather than being absent.
	s, got := dial(t, `{"p-1":{"id":"sc-1","name":"Размеры"},"p-2":null}`)
	byProduct, err := s.ProductScalesFor(context.Background(), false, []string{"p-1", "p-2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(byProduct) != 2 {
		t.Fatalf("got %+v", byProduct)
	}
	if byProduct["p-1"] == nil || byProduct["p-1"].Name != "Размеры" {
		t.Errorf("p-1 = %+v", byProduct["p-1"])
	}
	if got, ok := byProduct["p-2"]; !ok || got != nil {
		t.Errorf("p-2 must be present and nil, got %+v ok=%v", got, ok)
	}
	q := got2Query(got)
	if ids := q["productId"]; len(ids) != 2 {
		t.Errorf("productId = %v", ids)
	}
	if q.Get("includeDeletedProducts") != "false" {
		t.Errorf("includeDeletedProducts = %q", q.Get("includeDeletedProducts"))
	}
}

func TestProductScalesForWithNoProductIDsAsksForEverything(t *testing.T) {
	t.Parallel()

	// Omitting productId is legal and means every non-deleted product. It is not
	// an error, but it is the call that drags the whole catalogue over the wire.
	s, got := dial(t, `{}`)
	if _, err := s.ProductScalesFor(context.Background(), false, nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := got2Query(got)["productId"]; ok {
		t.Error("no productId must be sent, not an empty one")
	}
}

func TestSaveProductScaleCreatesADefinition(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `{"result":"SUCCESS","response":{"id":"sc-1","name":"Размеры"}}`)
	scale, err := s.SaveProductScale(context.Background(), ProductScaleDto{
		Name:         "Размеры",
		ProductSizes: []ProductSizeDto{{Name: "S", ShortName: "S", Default: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if scale.ID != "sc-1" {
		t.Fatalf("got %+v", scale)
	}
	if !strings.Contains(got.body, `"name":"Размеры"`) || !strings.Contains(got.body, `"shortName":"S"`) {
		t.Errorf("body was %s", got.body)
	}
	// deleted is read-only; echoing it back is rejected.
	if strings.Contains(got.body, "deleted") {
		t.Errorf("read-only deleted was sent: %s", got.body)
	}
}

func TestQueryProductScalesForPostsAForm(t *testing.T) {
	t.Parallel()

	s, got := dial(t, `{"p-1":null}`)
	byProduct, err := s.QueryProductScalesFor(context.Background(), true, []string{"p-1"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := byProduct["p-1"]; !ok || got != nil {
		t.Fatalf("p-1 = %+v ok=%v", got, ok)
	}
	if got.method != http.MethodPost || !strings.Contains(got.body, "productId=p-1") {
		t.Errorf("%s %q", got.method, got.body)
	}
	if !strings.Contains(got.body, "includeDeletedProducts=true") {
		t.Errorf("form was %q", got.body)
	}
}
