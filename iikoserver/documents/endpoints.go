package documents

// Documents: v1 XML dry run, v2 JSON store documents.
const (
	EndpointInventoryCheck       = "/api/documents/check/incomingInventory"  // XML, incomingInventoryValidationResult
	EndpointWriteoffList         = "/api/v2/documents/writeoff"              // GET list, POST create
	EndpointWriteoffByID         = "/api/v2/documents/writeoff/byId"         // JSON, bare object
	EndpointInternalTransferList = "/api/v2/documents/internalTransfer"      // GET list, POST create
	EndpointInternalTransferByID = "/api/v2/documents/internalTransfer/byId" // JSON, bare object

	EndpointWriteoffByNumber         = "/api/v2/documents/writeoff/byNumber"         // JSON, bare ARRAY
	EndpointInternalTransferByNumber = "/api/v2/documents/internalTransfer/byNumber" // JSON, bare ARRAY

	// v1 documents — XML in, XML out, answering <documentValidationResult>.
	// HTTP 200 is not success: a rejected document returns 200 with <valid>false</valid>.
	EndpointImportIncomingInvoice    = "/api/documents/import/incomingInvoice"
	EndpointImportOutgoingInvoice    = "/api/documents/import/outgoingInvoice"
	EndpointImportReturnedInvoice    = "/api/documents/import/returnedInvoice"
	EndpointImportProductionDocument = "/api/documents/import/productionDocument"
	EndpointImportSalesDocument      = "/api/documents/import/salesDocument"
	EndpointImportIncomingInventory  = "/api/documents/import/incomingInventory"

	EndpointUnprocessIncomingInvoice = "/api/documents/unprocess/incomingInvoice"
	EndpointUnprocessOutgoingInvoice = "/api/documents/unprocess/outgoingInvoice"

	EndpointExportIncomingInvoice         = "/api/documents/export/incomingInvoice"
	EndpointExportIncomingInvoiceByNumber = "/api/documents/export/incomingInvoice/byNumber"
	EndpointExportOutgoingInvoice         = "/api/documents/export/outgoingInvoice"
	EndpointExportOutgoingInvoiceByNumber = "/api/documents/export/outgoingInvoice/byNumber"
)
