package reports

// Reports — JSON.
const (
	EndpointOlap                 = "/api/v2/reports/olap"
	EndpointOlapColumns          = "/api/v2/reports/olap/columns"
	EndpointBalanceStores        = "/api/v2/reports/balance/stores"
	EndpointBalanceCounteragents = "/api/v2/reports/balance/counteragents"
	EndpointEgaisMarks           = "/api/v2/reports/egais/marks/list"

	// OLAP presets. The docs recommend fetching a preset's config and POSTing it
	// to /olap yourself rather than calling byPresetId, whose output shape shifts
	// between upgrades.
	EndpointOlapPresets       = "/api/v2/reports/olap/presets"
	EndpointOlapPresetsByType = "/api/v2/reports/olap/presets/{presetType}"
	EndpointOlapByPresetID    = "/api/v2/reports/olap/byPresetId/{presetId}"

	// v1 reports — XML, and their dates are DD.MM.YYYY, not ISO.
	EndpointOlapV1             = "/api/reports/olap"
	EndpointStoreOperations    = "/api/reports/storeOperations"
	EndpointStoreReportPresets = "/api/reports/storeReportPresets"
	EndpointProductExpense     = "/api/reports/productExpense"
	EndpointSalesReport        = "/api/reports/sales"
	EndpointMonthlyIncomePlan  = "/api/reports/monthlyIncomePlan"
	EndpointIngredientEntry    = "/api/reports/ingredientEntry"

	// Delivery reports. XML despite the docs labelling their samples JSON, and
	// their department filter is a literal-brace token that must be URL-encoded.
	EndpointDeliveryConsolidated     = "/api/reports/delivery/consolidated"
	EndpointDeliveryCouriers         = "/api/reports/delivery/couriers"
	EndpointDeliveryOrderCycle       = "/api/reports/delivery/orderCycle"
	EndpointDeliveryHalfHourDetailed = "/api/reports/delivery/halfHourDetailed"
	EndpointDeliveryRegions          = "/api/reports/delivery/regions"
	EndpointDeliveryLoyalty          = "/api/reports/delivery/loyalty"
)
