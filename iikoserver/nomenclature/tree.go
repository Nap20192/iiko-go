package nomenclature

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// TreeFilter narrows a group or category list. Unlike ProductFilter, both verbs
// of these two paths accept the same fields, so there is nothing to reject.
// The category list ignores Nums, Codes and ParentIDs — it is a flat list.
type TreeFilter struct {
	IncludeDeleted bool
	IDs            []string
	Nums           []string
	Codes          []string
	ParentIDs      []string
	RevisionFrom   *int64
}

func (f TreeFilter) values() url.Values {
	v := url.Values{
		"includeDeleted": {rest.BoolStr(f.IncludeDeleted)},
		"revisionFrom":   {"-1"},
	}
	if f.RevisionFrom != nil {
		v.Set("revisionFrom", strconv.FormatInt(*f.RevisionFrom, 10))
	}
	addAll(v, "ids", f.IDs)
	addAll(v, "nums", f.Nums)
	addAll(v, "codes", f.Codes)
	addAll(v, "parentIds", f.ParentIDs)
	return v
}

// ListGroups reads the nomenclature group tree over GET.
func (s *Service) ListGroups(ctx context.Context, f TreeFilter) ([]ProductGroupDto, error) {
	var out []ProductGroupDto
	if _, err := s.rest.GetV2(ctx, EndpointGroupList, f.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// QueryGroups reads the group tree over the POST twin of the same path.
func (s *Service) QueryGroups(ctx context.Context, f TreeFilter) ([]ProductGroupDto, error) {
	var out []ProductGroupDto
	if err := s.postForm(ctx, EndpointGroupList, f.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListCategories reads the user-defined product categories over GET. These are
// not tax or accounting categories.
func (s *Service) ListCategories(ctx context.Context, f TreeFilter) ([]EntityDto, error) {
	var out []EntityDto
	if _, err := s.rest.GetV2(ctx, EndpointCategoryList, f.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// QueryCategories reads the categories over the POST twin of the same path.
func (s *Service) QueryCategories(ctx context.Context, f TreeFilter) ([]EntityDto, error) {
	var out []EntityDto
	if err := s.postForm(ctx, EndpointCategoryList, f.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveGroup creates a nomenclature group. It changes production data.
func (s *Service) SaveGroup(ctx context.Context, g ProductGroupDto, generateNomenclatureCode, generateFastCode bool) (*ProductGroupDto, error) {
	q := url.Values{
		"generateNomenclatureCode": {rest.BoolStr(generateNomenclatureCode)},
		"generateFastCode":         {rest.BoolStr(generateFastCode)},
	}
	var out ProductGroupDto
	if _, err := s.rest.PostV2(ctx, EndpointGroupSave, q, g, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateGroup edits a group. It changes production data. Saving under a deleted
// parent is rejected upstream as PRODUCT_GROUP_DELETED.
func (s *Service) UpdateGroup(ctx context.Context, g ProductGroupDto, overrideFastCode, overrideNomenclatureCode bool) (*ProductGroupDto, error) {
	if strings.TrimSpace(g.ID) == "" {
		return nil, fmt.Errorf("id is required on update: without it the request addresses no element")
	}
	q := url.Values{
		"overrideFastCode":         {rest.BoolStr(overrideFastCode)},
		"overrideNomenclatureCode": {rest.BoolStr(overrideNomenclatureCode)},
	}
	var out ProductGroupDto
	if _, err := s.rest.PostV2(ctx, EndpointGroupUpdate, q, g, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ProductsAndGroups is what the combined delete and restore calls answer with.
type ProductsAndGroups struct {
	Products []ProductDto
	Groups   []ProductGroupDto
}

// DeleteProductsAndGroups soft-deletes products and groups in one call. It has
// to be one call: iiko refuses to delete a group while its children remain, and
// says so in a 409 whose plain text names the ones left behind.
func (s *Service) DeleteProductsAndGroups(ctx context.Context, productIDs, groupIDs []string) (*ProductsAndGroups, error) {
	return s.moveTree(ctx, EndpointGroupDelete, nil, productIDs, groupIDs)
}

// RestoreProductsAndGroups undoes a soft delete. A group cannot come back while
// its parent stays deleted, so restore its parent first.
func (s *Service) RestoreProductsAndGroups(ctx context.Context, productIDs, groupIDs []string, overrideNomenclatureCode bool) (*ProductsAndGroups, error) {
	q := url.Values{"overrideNomenclatureCode": {rest.BoolStr(overrideNomenclatureCode)}}
	return s.moveTree(ctx, EndpointGroupRestore, q, productIDs, groupIDs)
}

func (s *Service) moveTree(ctx context.Context, path string, q url.Values, productIDs, groupIDs []string) (*ProductsAndGroups, error) {
	if len(productIDs) == 0 && len(groupIDs) == 0 {
		return nil, fmt.Errorf("at least one product or group id is required")
	}
	body := ProductsAndGroupsDto{}
	var err error
	if body.Products, err = json.Marshal(idList(productIDs)); err != nil {
		return nil, err
	}
	if body.ProductGroups, err = json.Marshal(idList(groupIDs)); err != nil {
		return nil, err
	}

	var raw ProductsAndGroupsDto
	if _, err := s.rest.PostV2(ctx, path, q, body, &raw); err != nil {
		return nil, err
	}
	out := &ProductsAndGroups{}
	if len(raw.Products) > 0 {
		if err := json.Unmarshal(raw.Products, &out.Products); err != nil {
			return nil, fmt.Errorf("decode products: %w", err)
		}
	}
	if len(raw.ProductGroups) > 0 {
		if err := json.Unmarshal(raw.ProductGroups, &out.Groups); err != nil {
			return nil, fmt.Errorf("decode product groups: %w", err)
		}
	}
	return out, nil
}

// categoryRef is the bare {id} body delete and restore take. EntityDto's
// marshaller would add the required name, and whether iiko rejects the extra
// field is not documented either way.
type categoryRef struct {
	ID string `json:"id"`
}

// SaveCategory creates a product category. It changes production data.
func (s *Service) SaveCategory(ctx context.Context, name string) (*EntityDto, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is required and must not be whitespace")
	}
	return s.categoryWrite(ctx, EndpointCategorySave, EntityDto{Name: name})
}

// UpdateCategory renames a category.
func (s *Service) UpdateCategory(ctx context.Context, id, name string) (*EntityDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required on update")
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name is required and must not be whitespace")
	}
	return s.categoryWrite(ctx, EndpointCategoryUpdate, EntityDto{ID: id, Name: name})
}

// DeleteCategory soft-deletes a category. Deleting a deleted one is refused
// upstream in plain text.
func (s *Service) DeleteCategory(ctx context.Context, id string) (*EntityDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required")
	}
	return s.categoryWrite(ctx, EndpointCategoryDelete, categoryRef{ID: id})
}

// RestoreCategory undoes a soft delete. Restoring a live one is refused too.
func (s *Service) RestoreCategory(ctx context.Context, id string) (*EntityDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required")
	}
	return s.categoryWrite(ctx, EndpointCategoryRestore, categoryRef{ID: id})
}

func (s *Service) categoryWrite(ctx context.Context, path string, body any) (*EntityDto, error) {
	var out EntityDto
	if _, err := s.rest.PostV2(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
