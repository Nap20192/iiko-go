package suppliers

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the suppliers half of the iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// Supplier is a counterparty that delivers goods.
type Supplier struct {
	ID      string `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
}

// employees is the wire shape: suppliers come back as employee documents.
type employees struct {
	Items []struct {
		ID      string `xml:"id"`
		Code    string `xml:"code"`
		Name    string `xml:"name"`
		Deleted bool   `xml:"deleted"`
	} `xml:"employee"`
}

// List returns suppliers.
func (s *Service) List(ctx context.Context, includeDeleted bool) ([]Supplier, error) {
	q := url.Values{"includeDeleted": {rest.BoolStr(includeDeleted)}}
	return s.list(ctx, EndpointSuppliers, q, includeDeleted)
}

// searchableFields are the documented regex fields. id is absent on purpose:
// the docs list every other identifier, and searching by it matches nothing.
var searchableFields = map[string]bool{
	"name": true, "code": true, "phone": true, "cellPhone": true,
	"firstName": true, "middleName": true, "lastName": true,
	"email": true, "cardNumber": true, "taxpayerIdNumber": true,
}

// Search matches suppliers on the documented regex fields.
func (s *Service) Search(ctx context.Context, fields map[string]string) ([]Supplier, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("at least one search field is required: an unfiltered supplier list drags in every counterparty ever imported")
	}
	q := url.Values{}
	for k, v := range fields {
		if !searchableFields[k] {
			return nil, fmt.Errorf("suppliers cannot be searched by %q; iiko documents no such field (searching by id in particular is unsupported and silently matches nothing)", k)
		}
		q.Set(k, v)
	}
	return s.list(ctx, EndpointSuppliersSearch, q, true)
}

func (s *Service) list(ctx context.Context, path string, q url.Values, includeDeleted bool) ([]Supplier, error) {
	var raw employees
	if err := s.rest.GetXML(ctx, path, q, &raw); err != nil {
		return nil, err
	}
	out := make([]Supplier, 0, len(raw.Items))
	for _, e := range raw.Items {
		if e.Deleted && !includeDeleted {
			continue
		}
		out = append(out, Supplier{ID: e.ID, Code: e.Code, Name: e.Name, Deleted: e.Deleted})
	}
	return out, nil
}

// PriceListItem is one line of a supplier's price list.
type PriceListItem struct {
	ProductID string  `xml:"productId"`
	Price     float64 `xml:"price"`
}

type priceList struct {
	Items []PriceListItem `xml:"item"`
}

// Pricelist reads a supplier's prices, optionally as of a date.
//
// code is the supplier's CODE, not its UUID, and the date here is DD.MM.YYYY
// where neighbouring endpoints take ISO. Passing nil reads the latest list.
func (s *Service) Pricelist(ctx context.Context, code string, on *time.Time) ([]PriceListItem, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("supplier code is required: this endpoint takes the code, not the id")
	}
	path := strings.Replace(EndpointSupplierPricelist, "{code}", url.PathEscape(code), 1)
	q := url.Values{}
	if on != nil {
		q.Set("date", on.Format(rest.ReportV1))
	}
	var raw priceList
	if err := s.rest.GetXML(ctx, path, q, &raw); err != nil {
		return nil, err
	}
	return raw.Items, nil
}
