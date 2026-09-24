// Package inventory covers stock documents, their posting lifecycle and the inventory catalogues.
package inventory

import (
	"context"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// Service is the inventory half of iikoCloud.
type Service struct{ c *rest.Client }

// New wires a service onto a transport; the facade is the usual caller.
func New(c *rest.Client) *Service { return &Service{c: c} }

// AddIncomingInvoicePayment pay incoming invoice.
func (s *Service) AddIncomingInvoicePayment(ctx context.Context, req gen.PayRequest) error {
	return s.c.Post(ctx, addIncomingInvoicePayment, req, nil)
}

// AddOutgoingInvoicePayment pay outgoing invoice.
func (s *Service) AddOutgoingInvoicePayment(ctx context.Context, req gen.PayOutgoingInvoiceRequest) error {
	return s.c.Post(ctx, addOutgoingInvoicePayment, req, nil)
}

// CalculateCostings get cost prices for nomenclature items.
func (s *Service) CalculateCostings(ctx context.Context, req gen.GetCostPricesRequest) (*gen.GetCostPricesResponse, error) {
	return rest.Call[gen.GetCostPricesResponse](ctx, s.c, calculateCostings, req)
}

// CancelDisassembleDocument cancel disassemble document draft.
func (s *Service) CancelDisassembleDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.DisassembleDocumentSaveResponse, error) {
	return rest.Call[gen.DisassembleDocumentSaveResponse](ctx, s.c, cancelDisassembleDocument, req)
}

// CancelIncomingInventory cancel inventory draft.
func (s *Service) CancelIncomingInventory(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingInventorySaveResponse, error) {
	return rest.Call[gen.IncomingInventorySaveResponse](ctx, s.c, cancelIncomingInventory, req)
}

// CancelIncomingInvoice cancel incoming invoice draft.
func (s *Service) CancelIncomingInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingInvoiceSaveResponse, error) {
	return rest.Call[gen.IncomingInvoiceSaveResponse](ctx, s.c, cancelIncomingInvoice, req)
}

// CancelIncomingReturnedInvoice cancel incoming returned invoice draft.
func (s *Service) CancelIncomingReturnedInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingReturnedInvoiceSaveResponse, error) {
	return rest.Call[gen.IncomingReturnedInvoiceSaveResponse](ctx, s.c, cancelIncomingReturnedInvoice, req)
}

// CancelInternalTransfer cancel internal transfer act draft.
func (s *Service) CancelInternalTransfer(ctx context.Context, req gen.GetByIDRequest) (*gen.InternalTransferSaveResponse, error) {
	return rest.Call[gen.InternalTransferSaveResponse](ctx, s.c, cancelInternalTransfer, req)
}

// CancelOutgoingInvoice cancel outgoing invoice draft.
func (s *Service) CancelOutgoingInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.OutgoingInvoiceSaveResponse, error) {
	return rest.Call[gen.OutgoingInvoiceSaveResponse](ctx, s.c, cancelOutgoingInvoice, req)
}

// CancelProductionDocument cancel production document draft.
func (s *Service) CancelProductionDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.ProductionDocumentSaveResponse, error) {
	return rest.Call[gen.ProductionDocumentSaveResponse](ctx, s.c, cancelProductionDocument, req)
}

// CancelReturnedInvoice cancel returned invoice draft.
func (s *Service) CancelReturnedInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.ReturnedInvoiceSaveResponse, error) {
	return rest.Call[gen.ReturnedInvoiceSaveResponse](ctx, s.c, cancelReturnedInvoice, req)
}

// CancelSalesDocument cancel sales document draft.
func (s *Service) CancelSalesDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.SalesDocumentSaveResponse, error) {
	return rest.Call[gen.SalesDocumentSaveResponse](ctx, s.c, cancelSalesDocument, req)
}

// CancelTransformationDocument cancel transformation document draft.
func (s *Service) CancelTransformationDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.TransformationDocumentSaveResponse, error) {
	return rest.Call[gen.TransformationDocumentSaveResponse](ctx, s.c, cancelTransformationDocument, req)
}

// CancelWriteoffDocument cancel write-off document draft.
func (s *Service) CancelWriteoffDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.WriteoffDocumentSaveResponse, error) {
	return rest.Call[gen.WriteoffDocumentSaveResponse](ctx, s.c, cancelWriteoffDocument, req)
}

// CreateDisassembleDocument create disassemble document.
func (s *Service) CreateDisassembleDocument(ctx context.Context, req gen.DisassembleDocumentCreateRequest) error {
	return s.c.Post(ctx, createDisassembleDocument, req, nil)
}

// CreateIncomingInventory create inventory.
func (s *Service) CreateIncomingInventory(ctx context.Context, req gen.IncomingInventoryCreateRequest) error {
	return s.c.Post(ctx, createIncomingInventory, req, nil)
}

// CreateIncomingInvoice create incoming invoice.
func (s *Service) CreateIncomingInvoice(ctx context.Context, req gen.IncomingInvoiceRequest) error {
	return s.c.Post(ctx, createIncomingInvoice, req, nil)
}

// CreateIncomingReturnedInvoice create incoming returned invoice.
func (s *Service) CreateIncomingReturnedInvoice(ctx context.Context, req gen.IncomingReturnedInvoiceCreateRequest) error {
	return s.c.Post(ctx, createIncomingReturnedInvoice, req, nil)
}

// CreateInternalTransfer create internal transfer act.
func (s *Service) CreateInternalTransfer(ctx context.Context, req gen.InternalTransferCreateRequest) error {
	return s.c.Post(ctx, createInternalTransfer, req, nil)
}

// CreateOutgoingInvoice create outgoing invoice.
func (s *Service) CreateOutgoingInvoice(ctx context.Context, req gen.OutgoingInvoiceRequest) error {
	return s.c.Post(ctx, createOutgoingInvoice, req, nil)
}

// CreateProductionDocument create production document.
func (s *Service) CreateProductionDocument(ctx context.Context, req gen.ProductionDocumentCreateRequest) error {
	return s.c.Post(ctx, createProductionDocument, req, nil)
}

// CreateReturnedInvoice create returned invoice.
func (s *Service) CreateReturnedInvoice(ctx context.Context, req gen.ReturnedInvoiceCreateRequest) error {
	return s.c.Post(ctx, createReturnedInvoice, req, nil)
}

// CreateSalesDocument create sales document.
func (s *Service) CreateSalesDocument(ctx context.Context, req gen.SalesDocumentCreateRequest) error {
	return s.c.Post(ctx, createSalesDocument, req, nil)
}

// CreateTransformationDocument create transformation document.
func (s *Service) CreateTransformationDocument(ctx context.Context, req gen.TransformationDocumentCreateRequest) error {
	return s.c.Post(ctx, createTransformationDocument, req, nil)
}

// CreateWriteoffDocument create write-off document.
func (s *Service) CreateWriteoffDocument(ctx context.Context, req gen.WriteoffDocumentCreateRequest) error {
	return s.c.Post(ctx, createWriteoffDocument, req, nil)
}

// GetAccountingCategories get accounting category by ID.
func (s *Service) GetAccountingCategories(ctx context.Context, req gen.AccountingCategoryGetByIDRequest) (*gen.AccountingCategory, error) {
	return rest.Call[gen.AccountingCategory](ctx, s.c, getAccountingCategories, req)
}

// GetConceptions get conception by ID.
func (s *Service) GetConceptions(ctx context.Context, req gen.ConceptionGetByIDRequest) (*gen.Conception, error) {
	return rest.Call[gen.Conception](ctx, s.c, getConceptions, req)
}

// GetDisassembleDocument get disassemble document by identifier.
func (s *Service) GetDisassembleDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.DisassembleDocumentGetResponse, error) {
	return rest.Call[gen.DisassembleDocumentGetResponse](ctx, s.c, getDisassembleDocument, req)
}

// GetIncomingInventory get inventory.
func (s *Service) GetIncomingInventory(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingInventoryGetResponse, error) {
	return rest.Call[gen.IncomingInventoryGetResponse](ctx, s.c, getIncomingInventory, req)
}

// GetIncomingInvoice get incoming invoice by identifier.
func (s *Service) GetIncomingInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingInvoice, error) {
	return rest.Call[gen.IncomingInvoice](ctx, s.c, getIncomingInvoice, req)
}

// GetIncomingReturnedInvoice get incoming returned invoice by identifier.
func (s *Service) GetIncomingReturnedInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingReturnedInvoiceGetResponse, error) {
	return rest.Call[gen.IncomingReturnedInvoiceGetResponse](ctx, s.c, getIncomingReturnedInvoice, req)
}

// GetInternalTransfer get internal transfer act by identifier.
func (s *Service) GetInternalTransfer(ctx context.Context, req gen.GetByIDRequest) (*gen.InternalTransferGetResponse, error) {
	return rest.Call[gen.InternalTransferGetResponse](ctx, s.c, getInternalTransfer, req)
}

// GetMeasureUnits get measure unit by ID.
func (s *Service) GetMeasureUnits(ctx context.Context, req gen.MeasureUnitGetByIDRequest) (*gen.MeasureUnit, error) {
	return rest.Call[gen.MeasureUnit](ctx, s.c, getMeasureUnits, req)
}

// GetOrganizationsTree get terminal groups list.
func (s *Service) GetOrganizationsTree(ctx context.Context, req gen.OrganizationRequest) ([]gen.JurPerson, error) {
	return rest.CallList[gen.JurPerson](ctx, s.c, getOrganizationsTree, req)
}

// GetOutgoingInvoice get outgoing invoice by ID.
func (s *Service) GetOutgoingInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.OutgoingInvoice, error) {
	return rest.Call[gen.OutgoingInvoice](ctx, s.c, getOutgoingInvoice, req)
}

// GetPaymentTypes get payment type by ID.
func (s *Service) GetPaymentTypes(ctx context.Context, req gen.PaymentTypeGetByIDRequest) (*gen.PaymentType, error) {
	return rest.Call[gen.PaymentType](ctx, s.c, getPaymentTypes, req)
}

// GetProductionDocument get production document.
func (s *Service) GetProductionDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.ProductionDocumentGetResponse, error) {
	return rest.Call[gen.ProductionDocumentGetResponse](ctx, s.c, getProductionDocument, req)
}

// GetReturnedInvoice get returned invoice by identifier.
func (s *Service) GetReturnedInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.ReturnedInvoiceGetResponse, error) {
	return rest.Call[gen.ReturnedInvoiceGetResponse](ctx, s.c, getReturnedInvoice, req)
}

// GetSalesDocument get sales document.
func (s *Service) GetSalesDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.SalesDocumentGetResponse, error) {
	return rest.Call[gen.SalesDocumentGetResponse](ctx, s.c, getSalesDocument, req)
}

// GetTransformationDocument get transformation document.
func (s *Service) GetTransformationDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.TransformationDocumentGetResponse, error) {
	return rest.Call[gen.TransformationDocumentGetResponse](ctx, s.c, getTransformationDocument, req)
}

// GetWriteoffDocument get write-off document by identifier.
func (s *Service) GetWriteoffDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.WriteoffDocumentGetResponse, error) {
	return rest.Call[gen.WriteoffDocumentGetResponse](ctx, s.c, getWriteoffDocument, req)
}

// ListAccountingCategories get accounting categories list.
func (s *Service) ListAccountingCategories(ctx context.Context, req gen.AccountingCategoryListRequest) (*gen.AccountingCategoryListResponse, error) {
	return rest.Call[gen.AccountingCategoryListResponse](ctx, s.c, listAccountingCategories, req)
}

// ListConceptions get conceptions list.
func (s *Service) ListConceptions(ctx context.Context, req gen.ConceptionListRequest) (*gen.ConceptionListResponse, error) {
	return rest.Call[gen.ConceptionListResponse](ctx, s.c, listConceptions, req)
}

// ListCounteragents get counteragents list.
func (s *Service) ListCounteragents(ctx context.Context, req gen.GetCounteragentsRequest) (*gen.GetCounteragentsResponse, error) {
	return rest.Call[gen.GetCounteragentsResponse](ctx, s.c, listCounteragents, req)
}

// ListDisassembleDocument export disassemble documents.
func (s *Service) ListDisassembleDocument(ctx context.Context, req gen.ListRequest) ([]gen.DisassembleDocumentListItem, error) {
	return rest.CallList[gen.DisassembleDocumentListItem](ctx, s.c, listDisassembleDocument, req)
}

// ListIncomingInventory export inventories.
func (s *Service) ListIncomingInventory(ctx context.Context, req gen.ListRequest) ([]gen.IncomingInventoryListItem, error) {
	return rest.CallList[gen.IncomingInventoryListItem](ctx, s.c, listIncomingInventory, req)
}

// ListIncomingInvoice export incoming invoices.
func (s *Service) ListIncomingInvoice(ctx context.Context, req gen.ListRequest) ([]gen.IncomingInvoice, error) {
	return rest.CallList[gen.IncomingInvoice](ctx, s.c, listIncomingInvoice, req)
}

// ListIncomingReturnedInvoice export incoming returned invoices.
func (s *Service) ListIncomingReturnedInvoice(ctx context.Context, req gen.ListRequest) ([]gen.IncomingReturnedInvoiceListItem, error) {
	return rest.CallList[gen.IncomingReturnedInvoiceListItem](ctx, s.c, listIncomingReturnedInvoice, req)
}

// ListInternalTransfer export internal transfer acts.
func (s *Service) ListInternalTransfer(ctx context.Context, req gen.ListRequest) ([]gen.InternalTransferListItem, error) {
	return rest.CallList[gen.InternalTransferListItem](ctx, s.c, listInternalTransfer, req)
}

// ListMeasureUnits get measure units list.
func (s *Service) ListMeasureUnits(ctx context.Context, req gen.MeasureUnitListRequest) (*gen.MeasureUnitListResponse, error) {
	return rest.Call[gen.MeasureUnitListResponse](ctx, s.c, listMeasureUnits, req)
}

// ListOrganizationsSettings get corporation settings.
func (s *Service) ListOrganizationsSettings(ctx context.Context, req gen.OrganizationRequest) (*gen.CorporationSettings, error) {
	return rest.Call[gen.CorporationSettings](ctx, s.c, listOrganizationsSettings, req)
}

// ListOutgoingInvoice export outgoing invoices.
func (s *Service) ListOutgoingInvoice(ctx context.Context, req gen.ListRequest) ([]gen.OutgoingInvoice, error) {
	return rest.CallList[gen.OutgoingInvoice](ctx, s.c, listOutgoingInvoice, req)
}

// ListPaymentTypes get payment types list.
func (s *Service) ListPaymentTypes(ctx context.Context, req gen.PaymentTypeListRequest) (*gen.PaymentTypeListResponse, error) {
	return rest.Call[gen.PaymentTypeListResponse](ctx, s.c, listPaymentTypes, req)
}

// ListProductionDocument export production documents.
func (s *Service) ListProductionDocument(ctx context.Context, req gen.ListRequest) ([]gen.ProductionDocumentListItem, error) {
	return rest.CallList[gen.ProductionDocumentListItem](ctx, s.c, listProductionDocument, req)
}

// ListReturnedInvoice export returned invoices.
func (s *Service) ListReturnedInvoice(ctx context.Context, req gen.ListRequest) ([]gen.ReturnedInvoiceListItem, error) {
	return rest.CallList[gen.ReturnedInvoiceListItem](ctx, s.c, listReturnedInvoice, req)
}

// ListSalesDocument export sales documents.
func (s *Service) ListSalesDocument(ctx context.Context, req gen.ListRequest) ([]gen.SalesDocumentListItem, error) {
	return rest.CallList[gen.SalesDocumentListItem](ctx, s.c, listSalesDocument, req)
}

// ListStockBalance get stock balances by stores.
func (s *Service) ListStockBalance(ctx context.Context, req gen.StockBalanceListRequest) (*gen.StockBalanceListResponse, error) {
	return rest.Call[gen.StockBalanceListResponse](ctx, s.c, listStockBalance, req)
}

// ListStores get stores list.
func (s *Service) ListStores(ctx context.Context, req gen.StoresListRequest) (*gen.StoreListResponse, error) {
	return rest.Call[gen.StoreListResponse](ctx, s.c, listStores, req)
}

// ListTransformationDocument list transformation documents.
func (s *Service) ListTransformationDocument(ctx context.Context, req gen.ListRequest) ([]gen.TransformationDocumentListItem, error) {
	return rest.CallList[gen.TransformationDocumentListItem](ctx, s.c, listTransformationDocument, req)
}

// ListWriteoffDocument export write-off documents.
func (s *Service) ListWriteoffDocument(ctx context.Context, req gen.ListRequest) ([]gen.WriteoffDocumentListItem, error) {
	return rest.CallList[gen.WriteoffDocumentListItem](ctx, s.c, listWriteoffDocument, req)
}

// PostDisassembleDocument post disassemble document.
func (s *Service) PostDisassembleDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.DisassembleDocumentSaveResponse, error) {
	return rest.Call[gen.DisassembleDocumentSaveResponse](ctx, s.c, postDisassembleDocument, req)
}

// PostIncomingInventory post inventory.
func (s *Service) PostIncomingInventory(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingInventorySaveResponse, error) {
	return rest.Call[gen.IncomingInventorySaveResponse](ctx, s.c, postIncomingInventory, req)
}

// PostIncomingInvoice post incoming invoice.
func (s *Service) PostIncomingInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingInvoiceSaveResponse, error) {
	return rest.Call[gen.IncomingInvoiceSaveResponse](ctx, s.c, postIncomingInvoice, req)
}

// PostIncomingReturnedInvoice post incoming returned invoice.
func (s *Service) PostIncomingReturnedInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingReturnedInvoiceSaveResponse, error) {
	return rest.Call[gen.IncomingReturnedInvoiceSaveResponse](ctx, s.c, postIncomingReturnedInvoice, req)
}

// PostInternalTransfer post internal transfer act.
func (s *Service) PostInternalTransfer(ctx context.Context, req gen.GetByIDRequest) (*gen.InternalTransferSaveResponse, error) {
	return rest.Call[gen.InternalTransferSaveResponse](ctx, s.c, postInternalTransfer, req)
}

// PostOutgoingInvoice post outgoing invoice.
func (s *Service) PostOutgoingInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.OutgoingInvoiceSaveResponse, error) {
	return rest.Call[gen.OutgoingInvoiceSaveResponse](ctx, s.c, postOutgoingInvoice, req)
}

// PostProductionDocument post production document.
func (s *Service) PostProductionDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.ProductionDocumentSaveResponse, error) {
	return rest.Call[gen.ProductionDocumentSaveResponse](ctx, s.c, postProductionDocument, req)
}

// PostReturnedInvoice post returned invoice.
func (s *Service) PostReturnedInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.ReturnedInvoiceSaveResponse, error) {
	return rest.Call[gen.ReturnedInvoiceSaveResponse](ctx, s.c, postReturnedInvoice, req)
}

// PostSalesDocument post sales document.
func (s *Service) PostSalesDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.SalesDocumentSaveResponse, error) {
	return rest.Call[gen.SalesDocumentSaveResponse](ctx, s.c, postSalesDocument, req)
}

// PostTransformationDocument post transformation document.
func (s *Service) PostTransformationDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.TransformationDocumentSaveResponse, error) {
	return rest.Call[gen.TransformationDocumentSaveResponse](ctx, s.c, postTransformationDocument, req)
}

// PostWriteoffDocument post write-off document.
func (s *Service) PostWriteoffDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.WriteoffDocumentSaveResponse, error) {
	return rest.Call[gen.WriteoffDocumentSaveResponse](ctx, s.c, postWriteoffDocument, req)
}

// SetIncomingInvoicePaymentDate set payment date for incoming invoice.
func (s *Service) SetIncomingInvoicePaymentDate(ctx context.Context, req gen.SetPaymentDateRequest) (*gen.SetPaymentDateResponse, error) {
	return rest.Call[gen.SetPaymentDateResponse](ctx, s.c, setIncomingInvoicePaymentDate, req)
}

// SetOutgoingInvoicePaymentDate set payment date for outgoing invoice.
func (s *Service) SetOutgoingInvoicePaymentDate(ctx context.Context, req gen.SetPaymentDateOutgoingRequest) (*gen.SetPaymentDateOutgoingResponse, error) {
	return rest.Call[gen.SetPaymentDateOutgoingResponse](ctx, s.c, setOutgoingInvoicePaymentDate, req)
}

// UnpostDisassembleDocument unpost disassemble document.
func (s *Service) UnpostDisassembleDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.DisassembleDocumentSaveResponse, error) {
	return rest.Call[gen.DisassembleDocumentSaveResponse](ctx, s.c, unpostDisassembleDocument, req)
}

// UnpostIncomingInventory unpost inventory.
func (s *Service) UnpostIncomingInventory(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingInventorySaveResponse, error) {
	return rest.Call[gen.IncomingInventorySaveResponse](ctx, s.c, unpostIncomingInventory, req)
}

// UnpostIncomingInvoice unpost incoming invoice.
func (s *Service) UnpostIncomingInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingInvoiceSaveResponse, error) {
	return rest.Call[gen.IncomingInvoiceSaveResponse](ctx, s.c, unpostIncomingInvoice, req)
}

// UnpostIncomingReturnedInvoice unpost incoming returned invoice.
func (s *Service) UnpostIncomingReturnedInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.IncomingReturnedInvoiceSaveResponse, error) {
	return rest.Call[gen.IncomingReturnedInvoiceSaveResponse](ctx, s.c, unpostIncomingReturnedInvoice, req)
}

// UnpostInternalTransfer unpost internal transfer act.
func (s *Service) UnpostInternalTransfer(ctx context.Context, req gen.GetByIDRequest) (*gen.InternalTransferSaveResponse, error) {
	return rest.Call[gen.InternalTransferSaveResponse](ctx, s.c, unpostInternalTransfer, req)
}

// UnpostOutgoingInvoice unpost outgoing invoice.
func (s *Service) UnpostOutgoingInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.OutgoingInvoiceSaveResponse, error) {
	return rest.Call[gen.OutgoingInvoiceSaveResponse](ctx, s.c, unpostOutgoingInvoice, req)
}

// UnpostProductionDocument unpost production document.
func (s *Service) UnpostProductionDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.ProductionDocumentSaveResponse, error) {
	return rest.Call[gen.ProductionDocumentSaveResponse](ctx, s.c, unpostProductionDocument, req)
}

// UnpostReturnedInvoice unpost returned invoice.
func (s *Service) UnpostReturnedInvoice(ctx context.Context, req gen.GetByIDRequest) (*gen.ReturnedInvoiceSaveResponse, error) {
	return rest.Call[gen.ReturnedInvoiceSaveResponse](ctx, s.c, unpostReturnedInvoice, req)
}

// UnpostSalesDocument unpost sales document.
func (s *Service) UnpostSalesDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.SalesDocumentSaveResponse, error) {
	return rest.Call[gen.SalesDocumentSaveResponse](ctx, s.c, unpostSalesDocument, req)
}

// UnpostTransformationDocument unpost transformation document.
func (s *Service) UnpostTransformationDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.TransformationDocumentSaveResponse, error) {
	return rest.Call[gen.TransformationDocumentSaveResponse](ctx, s.c, unpostTransformationDocument, req)
}

// UnpostWriteoffDocument unpost write-off document.
func (s *Service) UnpostWriteoffDocument(ctx context.Context, req gen.GetByIDRequest) (*gen.WriteoffDocumentSaveResponse, error) {
	return rest.Call[gen.WriteoffDocumentSaveResponse](ctx, s.c, unpostWriteoffDocument, req)
}

// UpdateDisassembleDocument edit disassemble document.
func (s *Service) UpdateDisassembleDocument(ctx context.Context, req gen.DisassembleDocumentUpdateRequest) (*gen.DisassembleDocumentSaveResponse, error) {
	return rest.Call[gen.DisassembleDocumentSaveResponse](ctx, s.c, updateDisassembleDocument, req)
}

// UpdateIncomingInventory edit inventory.
func (s *Service) UpdateIncomingInventory(ctx context.Context, req gen.IncomingInventoryUpdateRequest) (*gen.IncomingInventorySaveResponse, error) {
	return rest.Call[gen.IncomingInventorySaveResponse](ctx, s.c, updateIncomingInventory, req)
}

// UpdateIncomingInvoice edit incoming invoice.
func (s *Service) UpdateIncomingInvoice(ctx context.Context, req gen.IncomingInvoiceRequest) (*gen.IncomingInvoiceSaveResponse, error) {
	return rest.Call[gen.IncomingInvoiceSaveResponse](ctx, s.c, updateIncomingInvoice, req)
}

// UpdateIncomingReturnedInvoice edit incoming returned invoice.
func (s *Service) UpdateIncomingReturnedInvoice(ctx context.Context, req gen.IncomingReturnedInvoiceUpdateRequest) (*gen.IncomingReturnedInvoiceSaveResponse, error) {
	return rest.Call[gen.IncomingReturnedInvoiceSaveResponse](ctx, s.c, updateIncomingReturnedInvoice, req)
}

// UpdateInternalTransfer edit internal transfer act.
func (s *Service) UpdateInternalTransfer(ctx context.Context, req gen.InternalTransferUpdateRequest) (*gen.InternalTransferSaveResponse, error) {
	return rest.Call[gen.InternalTransferSaveResponse](ctx, s.c, updateInternalTransfer, req)
}

// UpdateOutgoingInvoice edit outgoing invoice.
func (s *Service) UpdateOutgoingInvoice(ctx context.Context, req gen.OutgoingInvoiceRequest) (*gen.OutgoingInvoiceSaveResponse, error) {
	return rest.Call[gen.OutgoingInvoiceSaveResponse](ctx, s.c, updateOutgoingInvoice, req)
}

// UpdateProductionDocument edit production document.
func (s *Service) UpdateProductionDocument(ctx context.Context, req gen.ProductionDocumentUpdateRequest) (*gen.ProductionDocumentSaveResponse, error) {
	return rest.Call[gen.ProductionDocumentSaveResponse](ctx, s.c, updateProductionDocument, req)
}

// UpdateReturnedInvoice edit returned invoice.
func (s *Service) UpdateReturnedInvoice(ctx context.Context, req gen.ReturnedInvoiceUpdateRequest) (*gen.ReturnedInvoiceSaveResponse, error) {
	return rest.Call[gen.ReturnedInvoiceSaveResponse](ctx, s.c, updateReturnedInvoice, req)
}

// UpdateSalesDocument edit sales document.
func (s *Service) UpdateSalesDocument(ctx context.Context, req gen.SalesDocumentUpdateRequest) (*gen.SalesDocumentSaveResponse, error) {
	return rest.Call[gen.SalesDocumentSaveResponse](ctx, s.c, updateSalesDocument, req)
}

// UpdateTransformationDocument edit transformation document.
func (s *Service) UpdateTransformationDocument(ctx context.Context, req gen.TransformationDocumentUpdateRequest) (*gen.TransformationDocumentSaveResponse, error) {
	return rest.Call[gen.TransformationDocumentSaveResponse](ctx, s.c, updateTransformationDocument, req)
}

// UpdateWriteoffDocument edit write-off document.
func (s *Service) UpdateWriteoffDocument(ctx context.Context, req gen.WriteoffDocumentUpdateRequest) (*gen.WriteoffDocumentSaveResponse, error) {
	return rest.Call[gen.WriteoffDocumentSaveResponse](ctx, s.c, updateWriteoffDocument, req)
}
