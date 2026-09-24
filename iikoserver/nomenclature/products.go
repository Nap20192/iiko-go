package nomenclature

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// ProductFilter narrows a product list. Codes and RevisionFrom exist only on the
// POST form of the same path, so ListProducts rejects them and QueryProducts
// takes them.
//
// An empty string in ParentIDs is the documented filter for "no parent":
// parentIds=g-1&parentIds= means "in g-1, or top level".
type ProductFilter struct {
	IncludeDeleted bool
	IDs            []string
	Nums           []string
	Types          []string
	CategoryIDs    []string
	ParentIDs      []string

	Codes        []string
	RevisionFrom *int64
}

func (f ProductFilter) values() url.Values {
	v := url.Values{"includeDeleted": {rest.BoolStr(f.IncludeDeleted)}}
	addAll(v, "ids", f.IDs)
	addAll(v, "nums", f.Nums)
	addAll(v, "types", f.Types)
	addAll(v, "categoryIds", f.CategoryIDs)
	addAll(v, "parentIds", f.ParentIDs)
	return v
}

// ListProducts reads the catalogue over GET. Requires the B_EN right.
func (s *Service) ListProducts(ctx context.Context, f ProductFilter) ([]ProductDto, error) {
	if len(f.Codes) > 0 || f.RevisionFrom != nil {
		return nil, fmt.Errorf("codes and revisionFrom are only accepted by the POST form of %s: use QueryProducts", EndpointProductsList)
	}
	var out []ProductDto
	if _, err := s.rest.GetV2(ctx, EndpointProductsList, f.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// QueryProducts reads the catalogue over the POST twin of the same path, which
// additionally accepts codes and revisionFrom for incremental sync.
func (s *Service) QueryProducts(ctx context.Context, f ProductFilter) ([]ProductDto, error) {
	form := f.values()
	addAll(form, "codes", f.Codes)
	if f.RevisionFrom != nil {
		form.Set("revisionFrom", strconv.FormatInt(*f.RevisionFrom, 10))
	}
	var out []ProductDto
	if err := s.postForm(ctx, EndpointProductsList, form, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveProduct creates a product. It changes production data.
//
// Both switches default to true upstream and are always sent: with
// generateNomenclatureCode false the caller owns num, and with generateFastCode
// false it owns code — silence would let iiko overwrite either.
func (s *Service) SaveProduct(ctx context.Context, p ProductDto, generateNomenclatureCode, generateFastCode bool) (*ProductDto, error) {
	if !generateNomenclatureCode && strings.TrimSpace(p.Num) == "" {
		return nil, fmt.Errorf("num is required when generateNomenclatureCode is false")
	}
	if err := validateMenuFields(p); err != nil {
		return nil, err
	}
	q := url.Values{
		"generateNomenclatureCode": {rest.BoolStr(generateNomenclatureCode)},
		"generateFastCode":         {rest.BoolStr(generateFastCode)},
	}
	var out ProductDto
	if _, err := s.rest.PostV2(ctx, EndpointProductsSave, q, p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateProduct edits an existing product. It changes production data.
//
// A product's type may only move within the same colour-grouped category; iiko
// enforces that and reports it as a v2 error.
func (s *Service) UpdateProduct(ctx context.Context, p ProductDto, overrideFastCode, overrideNomenclatureCode bool) (*ProductDto, error) {
	if strings.TrimSpace(p.ID) == "" {
		return nil, fmt.Errorf("id is required on update: without it the request addresses no element")
	}
	// The num rule belongs to save: update has overrideNomenclatureCode instead,
	// and the product already carries a num assigned when it was created.
	if err := validateMenuFields(p); err != nil {
		return nil, err
	}
	q := url.Values{
		"overrideFastCode":         {rest.BoolStr(overrideFastCode)},
		"overrideNomenclatureCode": {rest.BoolStr(overrideNomenclatureCode)},
	}
	var out ProductDto
	if _, err := s.rest.PostV2(ctx, EndpointProductsUpdate, q, p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteProducts soft-deletes products and returns them with deleted set.
// Deleting an already-deleted product is a 409 whose plain-text body names the
// offending ids; it is passed through as iiko wrote it.
func (s *Service) DeleteProducts(ctx context.Context, ids []string) ([]ProductDto, error) {
	if err := requireIDs(ids); err != nil {
		return nil, err
	}
	var out []ProductDto
	if _, err := s.rest.PostV2(ctx, EndpointProductsDelete, nil, idList(ids), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// RestoreProducts undoes a soft delete. overrideNomenclatureCode generates a new
// num when the restored one now collides with a live product.
func (s *Service) RestoreProducts(ctx context.Context, ids []string, overrideNomenclatureCode bool) ([]ProductDto, error) {
	if err := requireIDs(ids); err != nil {
		return nil, err
	}
	q := url.Values{"overrideNomenclatureCode": {rest.BoolStr(overrideNomenclatureCode)}}
	var out []ProductDto
	if _, err := s.rest.PostV2(ctx, EndpointProductsRestore, q, idList(ids), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// validateMenuFields checks the two menu rules the docs state outright. They are
// deterministic and local, unlike the date-relative rules elsewhere, so catching
// them here saves a round trip that would fail anyway.
func validateMenuFields(p ProductDto) error {
	if p.DefaultIncludeInMenu && strings.TrimSpace(p.PlaceType) == "" {
		return fmt.Errorf("placeType is required when defaultIncludeInMenu is true, whatever the product type")
	}
	if !p.DefaultIncludeInMenu && len(p.ExcludedSections) > 0 {
		return fmt.Errorf("excludedSections must be empty when defaultIncludeInMenu is false")
	}
	return nil
}

// productList is the <productDtoes> root of the legacy XML endpoints.
type productList struct {
	Product []ProductDtoXML `xml:"productDto"`
}

// ProductsXML reads the whole catalogue over the legacy v1 XML endpoint.
//
// The docs' own warning: this returns supplier goods too, and supplier goods
// cannot be deleted in iiko, so the tail of unused items never shrinks.
func (s *Service) ProductsXML(ctx context.Context, includeDeleted bool, revisionFrom int64) ([]ProductDtoXML, error) {
	q := url.Values{
		"includeDeleted": {rest.BoolStr(includeDeleted)},
		"revisionFrom":   {strconv.FormatInt(revisionFrom, 10)},
	}
	var out productList
	if err := s.rest.GetXML(ctx, EndpointProductsXML, q, &out); err != nil {
		return nil, err
	}
	return out.Product, nil
}

// ProductSearchXML filters the legacy XML search. Every field is a regular
// expression, not a literal: "Апельсин" matches anything containing it.
type ProductSearchXML struct {
	IncludeDeleted   bool
	Name             string
	Code             string
	MainUnit         string
	Num              string
	CookingPlaceType string
	ProductGroupType string // matched against PRODUCTS|MODIFIERS
	ProductType      string // matched against the ProductType vocabulary
}

// SearchProductsXML runs the legacy regex search. Empty filters are omitted
// rather than sent blank, since an empty regular expression matches everything.
func (s *Service) SearchProductsXML(ctx context.Context, f ProductSearchXML) ([]ProductDtoXML, error) {
	q := url.Values{"includeDeleted": {rest.BoolStr(f.IncludeDeleted)}}
	for k, v := range map[string]string{
		"name": f.Name, "code": f.Code, "mainUnit": f.MainUnit, "num": f.Num,
		"cookingPlaceType": f.CookingPlaceType, "productGroupType": f.ProductGroupType,
		"productType": f.ProductType,
	} {
		if v != "" {
			q.Set(k, v)
		}
	}
	var out productList
	if err := s.rest.GetXML(ctx, EndpointProductsSearchXML, q, &out); err != nil {
		return nil, err
	}
	return out.Product, nil
}
