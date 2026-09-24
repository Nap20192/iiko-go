package cashshifts

// Cash shifts — v2 JSON, iiko 5.4.
const (
	EndpointCashShiftsList    = "/api/v2/cashshifts/list"           // JSON, bare array
	EndpointCashShiftByID     = "/api/v2/cashshifts/byId/"          // JSON, bare object; + {sessionId}
	EndpointCashShiftPayments = "/api/v2/cashshifts/payments/list/" // JSON, bare object; + {sessionId}

	// This GET has a side effect: it returns the shift's acceptance document,
	// CREATING it if none exists yet.
	EndpointClosedSessionDocument = "/api/v2/cashshifts/closedSessionDocument/{id}"
	EndpointCashShiftSave         = "/api/v2/cashshifts/save" // accepts a shift; sumReal is editable for pay-outs only

	// Pay-ins and pay-outs. Writing needs F_APIO, reading the types needs B_APIO.
	EndpointAddPayOut     = "/api/v2/payInOuts/addPayOut"
	EndpointPayInOutTypes = "/api/v2/entities/payInOutTypes/list"

	EndpointPayrolls = "/api/v2/payrolls/list" // dateFrom and dateTo are both inclusive
)
