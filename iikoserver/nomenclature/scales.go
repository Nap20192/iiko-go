package nomenclature

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// scaleFilter is the query both verbs of the scale list share.
func scaleFilter(includeDeleted bool, ids []string) url.Values {
	v := url.Values{"includeDeleted": {rest.BoolStr(includeDeleted)}}
	addAll(v, "ids", ids)
	return v
}

// ListProductScales reads the size-scale definitions over GET.
func (s *Service) ListProductScales(ctx context.Context, includeDeleted bool, ids []string) ([]ProductScaleDto, error) {
	var out []ProductScaleDto
	if _, err := s.rest.GetV2(ctx, EndpointProductScales, scaleFilter(includeDeleted, ids), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// QueryProductScales reads the definitions over the POST twin of the same path.
func (s *Service) QueryProductScales(ctx context.Context, includeDeleted bool, ids []string) ([]ProductScaleDto, error) {
	var out []ProductScaleDto
	if err := s.postForm(ctx, EndpointProductScales, scaleFilter(includeDeleted, ids), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ProductScaleByID reads one scale definition, as a bare object.
func (s *Service) ProductScaleByID(ctx context.Context, productScaleID string) (*ProductScaleDto, error) {
	if strings.TrimSpace(productScaleID) == "" {
		return nil, fmt.Errorf("productScaleId is required")
	}
	var out ProductScaleDto
	path := EndpointProductScaleByID + url.PathEscape(productScaleID)
	if _, err := s.rest.GetV2(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SaveProductScale creates a scale definition. It changes production data.
func (s *Service) SaveProductScale(ctx context.Context, scale ProductScaleDto) (*ProductScaleDto, error) {
	var out ProductScaleDto
	if _, err := s.rest.PostV2(ctx, EndpointProductScaleSave, nil, scale, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateProductScale edits a scale definition. A productSizes entry sent without
// an id creates a new size rather than editing one, so ids must survive.
func (s *Service) UpdateProductScale(ctx context.Context, scale ProductScaleDto) (*ProductScaleDto, error) {
	if strings.TrimSpace(scale.ID) == "" {
		return nil, fmt.Errorf("id is required on update: without it the request addresses no scale")
	}
	var out ProductScaleDto
	if _, err := s.rest.PostV2(ctx, EndpointProductScaleUpdate, nil, scale, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteProductScales soft-deletes scale definitions.
func (s *Service) DeleteProductScales(ctx context.Context, ids []string) ([]ProductScaleDto, error) {
	return s.scaleIDWrite(ctx, EndpointProductScaleDelete, ids)
}

// RestoreProductScales undoes a soft delete.
func (s *Service) RestoreProductScales(ctx context.Context, ids []string) ([]ProductScaleDto, error) {
	return s.scaleIDWrite(ctx, EndpointProductScaleRestore, ids)
}

func (s *Service) scaleIDWrite(ctx context.Context, path string, ids []string) ([]ProductScaleDto, error) {
	if err := requireIDs(ids); err != nil {
		return nil, err
	}
	var out []ProductScaleDto
	if _, err := s.rest.PostV2(ctx, path, nil, idList(ids), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func bindingPath(productID string) string {
	return fmt.Sprintf(EndpointProductScaleBinding, url.PathEscape(productID))
}

// ProductScaleBinding reads the scale a product is bound to, with the product's
// own per-size disabled flags and write-off factors merged in. A product with no
// scale bound answers with a null payload, which comes back as a nil scale.
func (s *Service) ProductScaleBinding(ctx context.Context, productID string) (*ProductScaleDto, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("productId is required")
	}
	var out *ProductScaleDto
	if _, err := s.rest.GetV2(ctx, bindingPath(productID), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BindProductScale binds a scale to a product, or edits the binding. It changes
// production data. The body carries only the scale id and the per-size overrides;
// name, shortName, priority and default belong to the definition.
func (s *Service) BindProductScale(ctx context.Context, productID string, binding ProductScaleBindingDto) (*ProductScaleDto, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, fmt.Errorf("productId is required")
	}
	if strings.TrimSpace(binding.ID) == "" {
		return nil, fmt.Errorf("the scale id is required: a binding with no scale binds nothing")
	}
	var out ProductScaleDto
	if _, err := s.rest.PostV2(ctx, bindingPath(productID), nil, binding, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UnbindProductScale detaches the scale from a product and returns the scale's
// id. It does not delete the definition — DeleteProductScales does that.
//
// The docs give no example for this call and describe the response only as "the
// scale's UUID", so the body is returned as the plain string iiko sends.
func (s *Service) UnbindProductScale(ctx context.Context, productID string) (string, error) {
	if strings.TrimSpace(productID) == "" {
		return "", fmt.Errorf("productId is required")
	}
	data, err := s.rest.Do(ctx, http.MethodDelete, bindingPath(productID), nil, nil, "")
	if err != nil {
		return "", err
	}
	return strings.Trim(strings.TrimSpace(string(data)), `"`), nil
}

func batchFilter(includeDeletedProducts bool, productIDs []string) url.Values {
	v := url.Values{"includeDeletedProducts": {rest.BoolStr(includeDeletedProducts)}}
	addAll(v, "productId", productIDs)
	return v
}

// ProductScalesFor reads scale bindings for many products at once, keyed by
// product id. A product with no scale bound maps to nil.
//
// An empty productIDs asks for every non-deleted product, which is legal and is
// the call that drags the whole catalogue over the wire.
func (s *Service) ProductScalesFor(ctx context.Context, includeDeletedProducts bool, productIDs []string) (map[string]*ProductScaleDto, error) {
	out := map[string]*ProductScaleDto{}
	if _, err := s.rest.GetV2(ctx, EndpointProductScalesBatch, batchFilter(includeDeletedProducts, productIDs), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// QueryProductScalesFor is the POST twin of ProductScalesFor, for id lists too
// long to sit in a URL.
func (s *Service) QueryProductScalesFor(ctx context.Context, includeDeletedProducts bool, productIDs []string) (map[string]*ProductScaleDto, error) {
	out := map[string]*ProductScaleDto{}
	if err := s.postForm(ctx, EndpointProductScalesBatch, batchFilter(includeDeletedProducts, productIDs), &out); err != nil {
		return nil, err
	}
	return out, nil
}
