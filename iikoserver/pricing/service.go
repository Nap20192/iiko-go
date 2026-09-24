// Package pricing covers menu-change orders (приказы), the prices they set, and
// the price categories and period schedules those prices key on.
package pricing

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the pricing half of the iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// Price kinds returned by /v2/price. BASE holds for the whole requested
// interval; SCHEDULED holds only inside a period schedule within it.
const (
	PriceBase      = "BASE"
	PriceScheduled = "SCHEDULED"
)

// farFuture is the documented default upper bound of /v2/price, not a sentinel
// of ours: iiko substitutes exactly this date when dateTo is omitted.
const farFuture = "2500-01-01"

// ListPrices returns the prices menu-change orders set over [dateFrom, dateTo].
// A zero dateTo sends the documented 2500-01-01 rather than omitting it.
//
// The returned revision is the maximum available at request time; pass it back
// as revisionFrom to get only what changed. Use -1 for a full export.
func (s *Service) ListPrices(ctx context.Context, dateFrom, dateTo time.Time, departmentIDs []string,
	includeOutOfSale bool, priceType string, revisionFrom int64,
) ([]ProductPriceDto, *int64, error) {
	if dateFrom.IsZero() {
		return nil, nil, fmt.Errorf("dateFrom is required by /v2/price")
	}
	priceType = strings.TrimSpace(priceType)
	if priceType != "" && priceType != PriceBase && priceType != PriceScheduled {
		return nil, nil, fmt.Errorf("price type %q is not documented: use %s, %s, or empty for both",
			priceType, PriceBase, PriceScheduled)
	}

	q := url.Values{
		"dateFrom":         {dateFrom.Format(rest.QueryV2)},
		"dateTo":           {farFuture},
		"includeOutOfSale": {rest.BoolStr(includeOutOfSale)},
		"revisionFrom":     {strconv.FormatInt(revisionFrom, 10)},
	}
	if !dateTo.IsZero() {
		q.Set("dateTo", dateTo.Format(rest.QueryV2))
	}
	if priceType != "" {
		q.Set("type", priceType)
	}
	for _, id := range departmentIDs {
		q.Add("departmentId", id)
	}

	var out []ProductPriceDto
	rev, err := s.rest.GetV2(ctx, EndpointPrice, q, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, rev, nil
}

// Menu-change order statuses. An omitted status returns every one of them.
const (
	StatusNew       = "NEW"
	StatusProcessed = "PROCESSED"
	StatusDeleted   = "DELETED"
)

// ListMenuChanges returns menu-change orders dated inside [dateFrom, dateTo].
// Both bounds are required upstream. status may be empty for all statuses.
func (s *Service) ListMenuChanges(ctx context.Context, dateFrom, dateTo time.Time,
	status string, revisionFrom int64,
) ([]MenuChangeDocumentDto, *int64, error) {
	if dateFrom.IsZero() || dateTo.IsZero() {
		return nil, nil, fmt.Errorf("dateFrom and dateTo are both required by %s", EndpointMenuChange)
	}
	status = strings.TrimSpace(status)
	switch status {
	case "", StatusNew, StatusProcessed, StatusDeleted:
	default:
		return nil, nil, fmt.Errorf("status %q is not documented: use %s, %s, %s, or empty for all",
			status, StatusNew, StatusProcessed, StatusDeleted)
	}

	q := url.Values{
		"dateFrom":     {dateFrom.Format(rest.QueryV2)},
		"dateTo":       {dateTo.Format(rest.QueryV2)},
		"revisionFrom": {strconv.FormatInt(revisionFrom, 10)},
	}
	if status != "" {
		q.Set("status", status)
	}
	var out []MenuChangeDocumentDto
	rev, err := s.rest.GetV2(ctx, EndpointMenuChange, q, &out)
	if err != nil {
		return nil, nil, err
	}
	return out, rev, nil
}

// MenuChangeByID returns one order. The response is a bare object, not the
// v2 envelope the list on the same page uses.
func (s *Service) MenuChangeByID(ctx context.Context, id string) (*MenuChangeDocumentDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required")
	}
	var out MenuChangeDocumentDto
	if _, err := s.rest.GetV2(ctx, EndpointMenuChangeByID, url.Values{"id": {id}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MenuChangeByNumber returns every order carrying a document number; numbers
// repeat across years, so this is an array and not a single document.
func (s *Service) MenuChangeByNumber(ctx context.Context, documentNumber string) ([]MenuChangeDocumentDto, error) {
	if strings.TrimSpace(documentNumber) == "" {
		return nil, fmt.Errorf("documentNumber is required")
	}
	var out []MenuChangeDocumentDto
	if _, err := s.rest.GetV2(ctx, EndpointMenuChangeByNumber, url.Values{"documentNumber": {documentNumber}}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveMenuChange creates an order when doc.ID is empty and edits it otherwise.
// It changes production data: an order sets the prices a till will charge.
//
// iiko enforces the edit window itself and the rules are date-relative to the
// server, so they are not duplicated here: a NEW order is editable; a PROCESSED
// one is editable while its dateIncoming is today or later; a PROCESSED order
// that already started accepts only a dateTo of today or later. taxCategoryId
// and taxCategoryEnabled need licence module 21052802 and are rejected without it.
func (s *Service) SaveMenuChange(ctx context.Context, doc MenuChangeDocumentDto) (*MenuChangeDocumentDto, error) {
	var out MenuChangeDocumentDto
	if _, err := s.rest.PostV2(ctx, EndpointMenuChange, nil, doc, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPriceCategories returns the client price categories prices key on.
// ids narrows the result; an empty slice applies no id filter.
func (s *Service) ListPriceCategories(ctx context.Context, includeDeleted bool, ids []string,
	revisionFrom int64,
) ([]ClientPriceCategoryDto, *int64, error) {
	var out []ClientPriceCategoryDto
	rev, err := s.rest.GetV2(ctx, EndpointPriceCategories, entityQuery(includeDeleted, ids, revisionFrom), &out)
	if err != nil {
		return nil, nil, err
	}
	return out, rev, nil
}

// PriceCategoryByID returns one category, as a bare object.
func (s *Service) PriceCategoryByID(ctx context.Context, id string) (*ClientPriceCategoryDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required")
	}
	var out ClientPriceCategoryDto
	if _, err := s.rest.GetV2(ctx, EndpointPriceCategoryByID, url.Values{"id": {id}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPeriodSchedules returns the period schedules a scheduled order runs on.
func (s *Service) ListPeriodSchedules(ctx context.Context, includeDeleted bool, ids []string,
	revisionFrom int64,
) ([]PeriodScheduleDto, *int64, error) {
	var out []PeriodScheduleDto
	rev, err := s.rest.GetV2(ctx, EndpointPeriodSchedules, entityQuery(includeDeleted, ids, revisionFrom), &out)
	if err != nil {
		return nil, nil, err
	}
	return out, rev, nil
}

// PeriodScheduleByID returns one schedule, as a bare object.
func (s *Service) PeriodScheduleByID(ctx context.Context, id string) (*PeriodScheduleDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required")
	}
	var out PeriodScheduleDto
	if _, err := s.rest.GetV2(ctx, EndpointPeriodScheduleByID, url.Values{"id": {id}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// entityQuery is the shape both v2 entity lists here share; includeDeleted is
// always explicit because upstream defaults differ per endpoint.
func entityQuery(includeDeleted bool, ids []string, revisionFrom int64) url.Values {
	q := url.Values{
		"includeDeleted": {rest.BoolStr(includeDeleted)},
		"revisionFrom":   {strconv.FormatInt(revisionFrom, 10)},
	}
	for _, id := range ids {
		q.Add("id", id)
	}
	return q
}

// DaysOfWeek reads a period's days as weekdays. iiko numbers them
// 1=Monday..7=Sunday, which matches neither java.util.Calendar (1=Sunday) nor
// this API's own quickLabels (0=Monday..6=Sunday), so the mapping lives here
// rather than at every call site.
func DaysOfWeek(p PeriodScheduleItemDto) ([]time.Weekday, error) {
	out := make([]time.Weekday, 0, len(p.DaysOfWeek))
	for _, n := range p.DaysOfWeek {
		if n < 1 || n > 7 {
			return nil, fmt.Errorf("day of week %d is out of range: iiko periods use 1=Monday..7=Sunday", n)
		}
		out = append(out, time.Weekday(n%7))
	}
	return out, nil
}
