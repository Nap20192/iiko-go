package nomenclature

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the nomenclature half of the iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// postForm posts application/x-www-form-urlencoded, which is what every /v2 list
// in this domain wants from its POST twin — not the JSON body PostV2 sends.
func (s *Service) postForm(ctx context.Context, path string, form url.Values, out any) error {
	data, err := s.rest.Do(ctx, http.MethodPost, path, nil,
		strings.NewReader(form.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}
	_, err = rest.DecodeV2(data, out)
	return err
}

// idList is the {items:[{id}]} body the delete and restore calls share.
func idList(ids []string) IdListDto {
	items := make([]IdListItemDto, len(ids))
	for i, id := range ids {
		items[i] = IdListItemDto{ID: id}
	}
	return IdListDto{Items: items}
}

// addAll repeats a query or form key once per value. An empty value is kept on
// purpose: the docs' null-filter convention makes "parentIds=" mean "no parent".
func addAll(v url.Values, key string, values []string) {
	for _, s := range values {
		v.Add(key, s)
	}
}

func requireIDs(ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("at least one id is required")
	}
	return nil
}

// ProductType values. save accepts all but OUTER and PETROL, and a product's
// type may only change within the same colour-grouped category.
const (
	TypeGoods    = "GOODS"
	TypeDish     = "DISH"
	TypePrepared = "PREPARED"
	TypeService  = "SERVICE"
	TypeModifier = "MODIFIER"
	TypeOuter    = "OUTER"
	TypePetrol   = "PETROL"
	TypeRate     = "RATE"
)

func ProductTypes() []any {
	return []any{TypeGoods, TypeDish, TypePrepared, TypeService, TypeModifier, TypeRate}
}

// Product is the subset of ProductDto worth spending tokens on. The full DTO
// has ~30 fields; SEO text, image links and modifier schemas are not things an
// agent answering a question about stock or recipes can act on.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Num         string  `json:"num"`
	Code        string  `json:"code"`
	Type        string  `json:"type"`
	MainUnit    string  `json:"mainUnit"`
	ParentID    string  `json:"parent"`
	CategoryID  string  `json:"productCategoryId"`
	Deleted     bool    `json:"deleted"`
	Weight      float64 `json:"weight"`
	ColdLossPct float64 `json:"coldLossPercent"`
	HotLossPct  float64 `json:"hotLossPercent"`
}

// FindProducts lists nomenclature. Requires the B_EN right.
//
// Note the docs' own warning: this pulls supplier goods too, and supplier goods
// cannot be deleted in iiko, so an unfiltered call drags in a long tail of
// items nobody uses. Always pass a filter.
func (s *Service) FindProducts(ctx context.Context, search string, types []string, includeDeleted bool) ([]Product, error) {
	q := url.Values{"includeDeleted": {rest.BoolStr(includeDeleted)}}
	for _, t := range types {
		q.Add("types", t)
	}
	var all []Product
	if _, err := s.rest.GetV2(ctx, EndpointProductsList, q, &all); err != nil {
		return nil, err
	}
	if search == "" {
		return all, nil
	}
	needle := strings.ToLower(search)
	out := make([]Product, 0, 32)
	for _, p := range all {
		if strings.Contains(strings.ToLower(p.Name), needle) ||
			strings.Contains(strings.ToLower(p.Num), needle) ||
			strings.Contains(strings.ToLower(p.Code), needle) {
			out = append(out, p)
		}
	}
	return out, nil
}
