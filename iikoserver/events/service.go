// Package events covers the iikoServer event journal: reading it by time window
// or by revision, adding events to it, and its type metadata tree.
package events

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the event journal half of the iikoServer API. Every call needs
// licence module 2200 and right B_VTJ; rest names both in its 403 hint.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// MarshalXML supplies the root element name, which the catalog has no field for,
// and drops an empty <orderNums> wrapper. encoding/xml emits that wrapper even
// with omitempty, and the metadata filter documents no order numbers at all.
func (x EventsRequestDataDto) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type both EventsRequestDataDto
	type eventsOnly struct {
		Events []string `xml:"events>event"`
	}
	start.Name = xml.Name{Local: "eventsRequestData"}
	if len(x.OrderNums) == 0 {
		return e.EncodeElement(eventsOnly{Events: x.Events}, start)
	}
	return e.EncodeElement(both(x), start)
}

// eventsList carries the root element name encoding/xml needs on the way out.
type eventsList struct {
	XMLName xml.Name   `xml:"eventsList"`
	Event   []EventDto `xml:"event,omitempty"`
}

// ListEvents reads the journal over a time window. from defaults to the start
// of the current day upstream when zero; to is exclusive and unbounded when zero.
func (s *Service) ListEvents(ctx context.Context, from, to time.Time) (*EventsListDto, error) {
	q := url.Values{}
	if !from.IsZero() {
		q.Set("from_time", from.Format(rest.OlapMs))
	}
	if !to.IsZero() {
		q.Set("to_time", to.Format(rest.OlapMs))
	}
	var out EventsListDto
	if err := s.rest.GetXML(ctx, EndpointEvents, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEventsSince reads the journal by revision cursor. It sends no time bound
// at all: the docs forbid combining to_time with from_rev, and a time bound
// present alongside the cursor is what turns a poll back into a full re-read.
func (s *Service) ListEventsSince(ctx context.Context, fromRev int) (*EventsListDto, error) {
	var out EventsListDto
	q := url.Values{"from_rev": {strconv.Itoa(fromRev)}}
	if err := s.rest.GetXML(ctx, EndpointEvents, q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// NextRevision is the cursor for the following poll. <revision> means "returned
// up to and including this one", so the next request starts one past it.
func NextRevision(list *EventsListDto) int { return list.Revision + 1 }

// Dedup drops repeated events, keeping the first of each id. Redelivery under a
// new revision is not documented as impossible and the id is the documented key.
func Dedup(in []EventDto) []EventDto {
	seen := make(map[string]bool, len(in))
	out := make([]EventDto, 0, len(in))
	for _, e := range in {
		if e.ID != "" && seen[e.ID] {
			continue
		}
		seen[e.ID] = true
		out = append(out, e)
	}
	return out
}

// FilterEvents reads the journal filtered by event type and order number.
// It is a POST but changes nothing.
func (s *Service) FilterEvents(ctx context.Context, eventTypes, orderNums []string) (*EventsListDto, error) {
	var out EventsListDto
	body := EventsRequestDataDto{Events: eventTypes, OrderNums: orderNums}
	if err := s.rest.PostXML(ctx, EndpointEvents, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddEvents writes events to the journal and returns them with the ids the
// server assigned. Any id set by the caller is stripped: the server owns it.
func (s *Service) AddEvents(ctx context.Context, in []EventDto) (*EventsListDto, error) {
	body := eventsList{Event: make([]EventDto, len(in))}
	copy(body.Event, in)
	for i := range body.Event {
		body.Event[i].StripReadOnly()
	}
	var out EventsListDto
	if err := s.rest.PostXML(ctx, EndpointEventsAdd, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Metadata returns the event type tree. A type's id is the value that appears
// as an event's <type>.
func (s *Service) Metadata(ctx context.Context) (*GroupsListDto, error) {
	var out GroupsListDto
	if err := s.rest.GetXML(ctx, EndpointEventsMetadata, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// FilterMetadata returns the type tree narrowed to the given event types.
func (s *Service) FilterMetadata(ctx context.Context, eventTypes []string) (*GroupsListDto, error) {
	var out GroupsListDto
	if err := s.rest.PostXML(ctx, EndpointEventsMetadata, nil, EventsRequestDataDto{Events: eventTypes}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSessions returns till-shift information as raw XML. The docs describe the
// payload in prose only — no field table, no example — so there is nothing to
// decode it into that would not be invented.
func (s *Service) ListSessions(ctx context.Context, from, to time.Time) ([]byte, error) {
	q := url.Values{}
	if !from.IsZero() {
		q.Set("from_time", from.Format(rest.OlapMs))
	}
	if !to.IsZero() {
		q.Set("to_time", to.Format(rest.OlapMs))
	}
	return s.rest.Do(ctx, http.MethodGet, EndpointEventSessions, q, nil, "")
}
