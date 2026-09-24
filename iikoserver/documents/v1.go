package documents

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// v1 documents are XML in and XML out, and every import answers
// <documentValidationResult>. HTTP 200 is not success: a rejected document comes
// back 200 with <valid>false</valid> and a message written for a human.

// validationError turns a rejected document into an error carrying iiko's own
// wording, so the caller never mistakes a rejection for an empty result.
func validationError(res DocumentValidationResultDto) error {
	msg := strings.TrimSpace(res.ErrorMessage)
	if extra := strings.TrimSpace(res.AdditionalInfo); extra != "" {
		msg = strings.TrimSpace(msg + "\n" + extra)
	}
	if msg == "" {
		msg = "iiko rejected the document without a message"
	}
	return &rest.Error{
		Status: http.StatusOK,
		Body:   msg,
		Hint:   "iiko accepted the request but rejected the document (<valid>false</valid>). The message above is iiko's own and is written for a human — read it literally.",
	}
}

// importDoc posts a v1 document and checks the body, not the status.
func (s *Service) importDoc(ctx context.Context, endpoint string, doc any) (*DocumentValidationResultDto, error) {
	var res DocumentValidationResultDto
	if err := s.rest.PostXML(ctx, endpoint, nil, doc, &res); err != nil {
		return nil, err
	}
	if !res.Valid {
		return nil, validationError(res)
	}
	return &res, nil
}

// ImportIncomingInvoice posts a supplier delivery.
//
// Store placement: a document-level defaultStore requires the same store on
// every line, so one of the two must be set.
func (s *Service) ImportIncomingInvoice(ctx context.Context, doc IncomingInvoiceDto) (*DocumentValidationResultDto, error) {
	if doc.DefaultStore == "" && !everyLineHasStore(doc) {
		return nil, fmt.Errorf("incomingInvoice needs a defaultStore, or a store on every line: iiko rejects a partial placement with a 409")
	}
	return s.importDoc(ctx, EndpointImportIncomingInvoice, doc)
}

func everyLineHasStore(doc IncomingInvoiceDto) bool {
	if len(doc.Items) == 0 {
		return false
	}
	for _, it := range doc.Items {
		if it.Store == "" {
			return false
		}
	}
	return true
}

// ImportOutgoingInvoice posts a shipment out.
//
// Store placement here is document-level OR line-level, never both.
func (s *Service) ImportOutgoingInvoice(ctx context.Context, doc OutgoingInvoiceDto) (*DocumentValidationResultDto, error) {
	docLevel := doc.DefaultStoreID != "" || doc.DefaultStoreCode != ""
	if !docLevel && len(doc.Items) == 0 {
		return nil, fmt.Errorf("outgoingInvoice needs a default store or at least one line carrying its own store")
	}
	return s.importDoc(ctx, EndpointImportOutgoingInvoice, doc)
}

// ImportReturnedInvoice posts a return to a supplier.
//
// The original invoice's number and date are mandatory: without them iiko cannot
// tell which delivery is being returned.
func (s *Service) ImportReturnedInvoice(ctx context.Context, doc ReturnedInvoiceDto) (*DocumentValidationResultDto, error) {
	if strings.TrimSpace(doc.IncomingInvoiceNumber) == "" || strings.TrimSpace(doc.IncomingInvoiceDate) == "" {
		return nil, fmt.Errorf("returnedInvoice requires incomingInvoiceNumber and incomingInvoiceDate: they identify the delivery being returned")
	}
	return s.importDoc(ctx, EndpointImportReturnedInvoice, doc)
}

// ImportProductionDocument posts a preparation act.
//
// Placement is document-level only: storeFrom and storeTo, with no per-line store.
func (s *Service) ImportProductionDocument(ctx context.Context, doc ProductionDocumentDto) (*DocumentValidationResultDto, error) {
	if strings.TrimSpace(doc.StoreFrom) == "" || strings.TrimSpace(doc.StoreTo) == "" {
		return nil, fmt.Errorf("productionDocument requires storeFrom and storeTo: it has no per-line store field at all")
	}
	return s.importDoc(ctx, EndpointImportProductionDocument, doc)
}

// ImportSalesDocument posts a sales act. Placement is per-line only.
func (s *Service) ImportSalesDocument(ctx context.Context, doc SalesDocumentDto) (*DocumentValidationResultDto, error) {
	return s.importDoc(ctx, EndpointImportSalesDocument, doc)
}

// ImportIncomingInventory posts a real inventory count. Unlike CheckInventory
// this one persists.
func (s *Service) ImportIncomingInventory(ctx context.Context, doc InventoryDocument) (*InventoryValidation, error) {
	if len(doc.Items) == 0 {
		return nil, fmt.Errorf("an inventory needs at least one counted item")
	}
	if doc.StoreID == "" && doc.StoreCode == "" {
		return nil, fmt.Errorf("storeId or storeCode is required: iiko places an inventory document at the document level, never per line")
	}
	var out InventoryValidation
	if err := s.rest.PostXML(ctx, EndpointImportIncomingInventory, nil, doc, &out); err != nil {
		return nil, err
	}
	if !out.Valid {
		return nil, validationError(DocumentValidationResultDto{
			ErrorMessage: out.ErrorMessage, AdditionalInfo: out.AdditionalInfo,
		})
	}
	return &out, nil
}

// UnprocessIncomingInvoice un-posts a delivery, returning the stock it moved.
func (s *Service) UnprocessIncomingInvoice(ctx context.Context, number string) error {
	return s.unprocess(ctx, EndpointUnprocessIncomingInvoice, number)
}

// UnprocessOutgoingInvoice un-posts a shipment.
func (s *Service) UnprocessOutgoingInvoice(ctx context.Context, number string) error {
	return s.unprocess(ctx, EndpointUnprocessOutgoingInvoice, number)
}

func (s *Service) unprocess(ctx context.Context, endpoint, number string) error {
	if strings.TrimSpace(number) == "" {
		return fmt.Errorf("document number is required")
	}
	var res DocumentValidationResultDto
	if err := s.rest.PostXML(ctx, endpoint, url.Values{"documentNumber": {number}}, nil, &res); err != nil {
		return err
	}
	if !res.Valid {
		return validationError(res)
	}
	return nil
}

// ByNumberRequest fetches one document by its accounting number.
//
// The three fields interlock, and the docs are strict: CurrentYear is mandatory;
// true means the current year and forbids a window; false requires both bounds.
type ByNumberRequest struct {
	Number      string
	CurrentYear *bool
	From, To    *time.Time
}

func (r ByNumberRequest) query() (url.Values, error) {
	if strings.TrimSpace(r.Number) == "" {
		return nil, fmt.Errorf("number is required")
	}
	if r.CurrentYear == nil {
		return nil, fmt.Errorf("currentYear is required: iiko has no default for it")
	}
	q := url.Values{"number": {r.Number}, "currentYear": {rest.BoolStr(*r.CurrentYear)}}
	if *r.CurrentYear {
		if r.From != nil || r.To != nil {
			return nil, fmt.Errorf("currentYear=true must not carry from/to: it already means the current year")
		}
		return q, nil
	}
	if r.From == nil || r.To == nil {
		return nil, fmt.Errorf("currentYear=false requires both from and to")
	}
	q.Set("from", r.From.Format(rest.QueryV2))
	q.Set("to", r.To.Format(rest.QueryV2))
	return q, nil
}

type incomingInvoiceList struct {
	Items []IncomingInvoiceDto `xml:"document"`
}

type outgoingInvoiceList struct {
	Items []OutgoingInvoiceDto `xml:"document"`
}

// ExportIncomingInvoices lists supplier deliveries in a window. from and to are
// inclusive and ignore the time of day.
func (s *Service) ExportIncomingInvoices(ctx context.Context, from, to time.Time, supplierID string, revisionFrom int) ([]IncomingInvoiceDto, error) {
	q := url.Values{
		"from":         {from.Format(rest.QueryV2)},
		"to":           {to.Format(rest.QueryV2)},
		"revisionFrom": {fmt.Sprint(revisionFrom)},
	}
	if supplierID != "" {
		q.Set("supplierId", supplierID)
	}
	var out incomingInvoiceList
	if err := s.rest.GetXML(ctx, EndpointExportIncomingInvoice, q, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// ExportIncomingInvoiceByNumber fetches one delivery by its number.
func (s *Service) ExportIncomingInvoiceByNumber(ctx context.Context, req ByNumberRequest) ([]IncomingInvoiceDto, error) {
	q, err := req.query()
	if err != nil {
		return nil, err
	}
	var out incomingInvoiceList
	if err := s.rest.GetXML(ctx, EndpointExportIncomingInvoiceByNumber, q, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// ExportOutgoingInvoices lists shipments in a window.
func (s *Service) ExportOutgoingInvoices(ctx context.Context, from, to time.Time, supplierID string) ([]OutgoingInvoiceDto, error) {
	q := url.Values{"from": {from.Format(rest.QueryV2)}, "to": {to.Format(rest.QueryV2)}}
	if supplierID != "" {
		q.Set("supplierId", supplierID)
	}
	var out outgoingInvoiceList
	if err := s.rest.GetXML(ctx, EndpointExportOutgoingInvoice, q, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// ExportOutgoingInvoiceByNumber fetches one shipment by its number.
func (s *Service) ExportOutgoingInvoiceByNumber(ctx context.Context, req ByNumberRequest) ([]OutgoingInvoiceDto, error) {
	q, err := req.query()
	if err != nil {
		return nil, err
	}
	var out outgoingInvoiceList
	if err := s.rest.GetXML(ctx, EndpointExportOutgoingInvoiceByNumber, q, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// DocumentsByNumber fetches v2 store documents by accounting number.
//
// Unlike byId, which answers a bare object, this answers a bare ARRAY: one
// number can match several documents across years.
func (s *Service) DocumentsByNumber(ctx context.Context, kind DocumentKind, number string) ([]StoreDocument, error) {
	var endpoint string
	switch kind {
	case KindWriteoff:
		endpoint = EndpointWriteoffByNumber
	case KindInternalTransfer:
		endpoint = EndpointInternalTransferByNumber
	default:
		return nil, fmt.Errorf("unknown document kind %q; valid kinds: writeoff, internalTransfer", kind)
	}
	if strings.TrimSpace(number) == "" {
		return nil, fmt.Errorf("documentNumber is required")
	}
	var out []StoreDocument
	if _, err := s.rest.GetV2(ctx, endpoint, url.Values{"documentNumber": {number}}, &out); err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Kind = kind
	}
	return out, nil
}
