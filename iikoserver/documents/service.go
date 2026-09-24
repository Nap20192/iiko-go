package documents

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the store documents half of the iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// Inventory documents, v1 XML. Field names and their quirks come from
// research/dtos.yaml (IncomingInventoryDto, IncomingInventoryItemDto,
// IncomingInventoryValidationResultDto), which was built from the XSDs and
// worked examples on the iiko docs pages.
//
// Only the dry-run check is implemented here. The import twin persists a real
// document and is deliberately absent until the dry run has been exercised
// against a live stand.

// InventoryDocument is the request body of documents/check/incomingInventory.
// The XML root element is <document>, per the XSD and every worked example.
type InventoryDocument struct {
	XMLName xml.Name `xml:"document"`

	DocumentNumber string `xml:"documentNumber,omitempty"`
	DateIncoming   string `xml:"dateIncoming,omitempty"`
	Status         string `xml:"status,omitempty"`
	StoreID        string `xml:"storeId,omitempty"`
	StoreCode      string `xml:"storeCode,omitempty"`
	Comment        string `xml:"comment,omitempty"`

	Items []InventoryItem `xml:"items>item"`
}

// InventoryItem is one counted line.
//
// There is no plain <amount> element in the XSD: the counted quantity goes in
// amountContainer, which is in the units of containerId/containerCode when one
// is given and in the product's base units when neither is. Both forms appear
// in the docs' own example.
type InventoryItem struct {
	ProductID       string  `xml:"productId,omitempty"`
	ProductArticle  string  `xml:"productArticle,omitempty"`
	ContainerID     string  `xml:"containerId,omitempty"`
	ContainerCode   string  `xml:"containerCode,omitempty"`
	AmountContainer float64 `xml:"amountContainer"`
	Comment         string  `xml:"comment,omitempty"`
}

// IDCodeName is the idCodeNameDto shape iiko uses for referenced entities.
type IDCodeName struct {
	ID   string `xml:"id"`
	Code string `xml:"code"`
	Name string `xml:"name"`
}

// InventoryValidation is the response. The XSD names the root element
// <document>; the worked example emits <incomingInventoryValidationResult>.
// XMLName is left off so encoding/xml accepts either.
type InventoryValidation struct {
	Valid                bool   `xml:"valid"`
	Warning              bool   `xml:"warning"`
	DocumentNumber       string `xml:"documentNumber"`
	OtherSuggestedNumber string `xml:"otherSuggestedNumber"`
	ErrorMessage         string `xml:"errorMessage"`
	AdditionalInfo       string `xml:"additionalInfo"`

	Store IDCodeName `xml:"store"`
	Date  string     `xml:"date"`

	Items []InventoryResultItem `xml:"items>item"`
}

// InventoryResultItem is the per-product surplus/shortage breakdown.
//
// expectedAmount carries up to 9 decimals where documents themselves allow 3.
// differenceAmount and differenceSum are 0 for a non-posted document, so a dry
// run reports expected vs actual and leaves the difference columns at zero.
type InventoryResultItem struct {
	Product          IDCodeName `xml:"product"`
	ExpectedAmount   float64    `xml:"expectedAmount"`
	ExpectedSum      float64    `xml:"expectedSum"`
	ActualAmount     float64    `xml:"actualAmount"`
	DifferenceAmount float64    `xml:"differenceAmount"`
	DifferenceSum    float64    `xml:"differenceSum"`
}

// InventoryDate formats t for an inventory document's dateIncoming. The XSD
// accepts yyyy-MM-dd'T'HH:mm:ss and yyyy-MM-dd; the dotted form is documented
// as "not recommended".
func InventoryDate(t time.Time) string { return t.Format(rest.Stamp) }

// CheckInventory dry-runs an inventory count: it returns what the surplus and
// shortage would be and saves nothing.
//
// HTTP 200 is not success here. A rejected document comes back 200 with
// <valid>false</valid> and an errorMessage written for a human, so the status
// code alone must never be trusted.
func (s *Service) CheckInventory(ctx context.Context, doc InventoryDocument) (*InventoryValidation, error) {
	if len(doc.Items) == 0 {
		return nil, fmt.Errorf("an inventory check needs at least one counted item")
	}
	if doc.StoreID == "" && doc.StoreCode == "" {
		return nil, fmt.Errorf("storeId or storeCode is required: iiko places an inventory document at the document level, never per line")
	}

	var out InventoryValidation
	if err := s.rest.PostXML(ctx, EndpointInventoryCheck, nil, doc, &out); err != nil {
		return nil, err
	}
	if !out.Valid {
		msg := strings.TrimSpace(out.ErrorMessage)
		if extra := strings.TrimSpace(out.AdditionalInfo); extra != "" {
			msg = strings.TrimSpace(msg + "\n" + extra)
		}
		if msg == "" {
			msg = "iiko rejected the document without a message"
		}
		return nil, &rest.Error{
			Status: http.StatusOK,
			Body:   msg,
			Hint:   "iiko accepted the request but rejected the document (<valid>false</valid>). The message above is iiko's own and is written for a human — read it literally.",
		}
	}
	return &out, nil
}

// ---------------------------------------------------------------------------
// Store documents, v2 JSON, iiko 7.9.3.
//
// Read-only here. Creating either document type is a write and belongs behind
// internal/writes, so only the list and byId shapes live in this file.
//
// Envelope asymmetry is per method: list answers the {result,errors,response,
// revision} envelope, byId answers the bare object. DecodeV2 absorbs both.

// DocumentStatus values shared by v2 store documents.
const (
	DocNew       = "NEW"
	DocProcessed = "PROCESSED"
	DocDeleted   = "DELETED"
)

// DocumentStatuses is the documented vocabulary for the status filter.
func DocumentStatuses() []any { return []any{DocNew, DocProcessed, DocDeleted} }

// DocumentKind selects which v2 store document list to read. These are separate
// endpoints upstream with different mandatory store fields, not a filter on one.
type DocumentKind string

const (
	KindWriteoff         DocumentKind = "writeoff"
	KindInternalTransfer DocumentKind = "internalTransfer"
)

// DocumentKinds is the vocabulary a caller may pass.
func DocumentKinds() []any { return []any{string(KindWriteoff), string(KindInternalTransfer)} }

// StoreDocument is the shared read shape of a writeoff and an internal
// transfer. The two differ only in store placement: a writeoff has StoreID and
// AccountID, a transfer has StoreFromID and StoreToID. Reading them through one
// type keeps one tool able to answer "what left the store last week".
//
// Num, MeasureUnitID and Cost on the lines are server-computed; they are here
// because they are useful to read, and must be stripped before any resend.
type StoreDocument struct {
	Kind DocumentKind `json:"-"` // set by the client, not by iiko

	ID             string `json:"id"`
	DateIncoming   string `json:"dateIncoming"`
	DocumentNumber string `json:"documentNumber"`
	Status         string `json:"status"`
	ConceptionID   string `json:"conceptionId"`
	Comment        string `json:"comment"`

	StoreID     string `json:"storeId"`     // writeoff
	AccountID   string `json:"accountId"`   // writeoff
	StoreFromID string `json:"storeFromId"` // internal transfer
	StoreToID   string `json:"storeToId"`   // internal transfer

	Items []StoreDocumentItem `json:"items"`
}

// StoreDocumentItem is one document line.
type StoreDocumentItem struct {
	Num           int     `json:"num"`
	ProductID     string  `json:"productId"`
	ProductSizeID string  `json:"productSizeId"`
	Amount        float64 `json:"amount"`
	AmountFactor  float64 `json:"amountFactor"`
	MeasureUnitID string  `json:"measureUnitId"`
	ContainerID   string  `json:"containerId"`
	Cost          float64 `json:"cost"`
}

// documentEndpoints maps a kind to its list and byId paths, keeping both in
// endpoints.go rather than assembled inline.
func documentEndpoints(kind DocumentKind) (list, byID string) {
	switch kind {
	case KindWriteoff:
		return EndpointWriteoffList, EndpointWriteoffByID
	case KindInternalTransfer:
		return EndpointInternalTransferList, EndpointInternalTransferByID
	default:
		return "", ""
	}
}

// ListDocuments returns v2 store documents of one kind in [from, to].
//
// dateFrom and dateTo are both mandatory upstream. status is optional; when
// empty it is omitted rather than sent blank.
func (s *Service) ListDocuments(ctx context.Context, kind DocumentKind, from, to time.Time, status string) ([]StoreDocument, error) {
	list, _ := documentEndpoints(kind)
	if list == "" {
		return nil, fmt.Errorf("unknown document kind %q; valid kinds: writeoff, internalTransfer", kind)
	}
	q := url.Values{
		"dateFrom": {from.Format(rest.QueryV2)},
		"dateTo":   {to.Format(rest.QueryV2)},
	}
	if s := strings.TrimSpace(status); s != "" {
		q.Set("status", s)
	}
	var out []StoreDocument
	if _, err := s.rest.GetV2(ctx, list, q, &out); err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Kind = kind
	}
	return out, nil
}

// DocumentByID returns one v2 store document. The response is a bare object,
// unlike the enveloped list.
func (s *Service) DocumentByID(ctx context.Context, kind DocumentKind, id string) (*StoreDocument, error) {
	_, byID := documentEndpoints(kind)
	if byID == "" {
		return nil, fmt.Errorf("unknown document kind %q; valid kinds: writeoff, internalTransfer", kind)
	}
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("document id is required")
	}
	var out StoreDocument
	if _, err := s.rest.GetV2(ctx, byID, url.Values{"id": {id}}, &out); err != nil {
		return nil, err
	}
	out.Kind = kind
	return &out, nil
}
