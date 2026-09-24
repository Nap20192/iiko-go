package pricing

// Menu-change orders, the prices they set, and the reference books those
// prices key on. All v2 JSON; the byId/byNumber pair is not enveloped.
const (
	EndpointMenuChange         = "/api/v2/documents/menuChange"          // JSON, v2-wrapped array; POST writes
	EndpointMenuChangeByID     = "/api/v2/documents/menuChange/byId"     // JSON, bare object
	EndpointMenuChangeByNumber = "/api/v2/documents/menuChange/byNumber" // JSON, bare array
	EndpointPrice              = "/api/v2/price"                         // JSON, v2-wrapped array
	EndpointPriceCategories    = "/api/v2/entities/priceCategories"      // JSON, v2-wrapped array
	EndpointPriceCategoryByID  = "/api/v2/entities/priceCategories/byId" // JSON, bare object
	EndpointPeriodSchedules    = "/api/v2/entities/periodSchedules"      // JSON, v2-wrapped array
	EndpointPeriodScheduleByID = "/api/v2/entities/periodSchedules/byId" // JSON, bare object
)
