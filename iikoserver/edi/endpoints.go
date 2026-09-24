package edi

// EDI order exchange. All XML. Every path takes the EDI system's GUID, which is
// configured in iikoOffice under Обмен данными → Системы EDI, so each constant
// is a template rather than a path.
//
// Eight of the ten writes are PUT, not POST. The docs' rendered HTTP blocks are
// unreliable here; the verbs below are the CSS-confirmed ones recorded in
// research/verbs-from-css.tsv, and TestVerbsMatchTheCatalog pins them.
const (
	EndpointInvoice          = "/api/edi/%s/invoice"           // XML; PUT
	EndpointOrdersAck        = "/api/edi/%s/orders/ack"        // XML; PUT, no body
	EndpointOrdersBySeller   = "/api/edi/%s/orders/bySeller"   // XML <ediMessageDtoes>; GET
	EndpointOrdersCreate     = "/api/edi/%s/orders/create"     // XML; POST — the one POST here
	EndpointOrdersList       = "/api/edi/%s/orders/list"       // XML <orderDtoes>; GET
	EndpointOrdersRegister   = "/api/edi/%s/orders/register"   // XML <orderDtoes>; PUT
	EndpointOrdersSend       = "/api/edi/%s/orders/send"       // XML <orderDtoes>; PUT
	EndpointOrdersUnregister = "/api/edi/%s/orders/unregister" // XML <orderDtoes>; PUT
	EndpointOrdersUpdate     = "/api/edi/%s/orders/update"     // XML; PUT
	EndpointResponse         = "/api/edi/%s/response"          // XML; PUT
)
