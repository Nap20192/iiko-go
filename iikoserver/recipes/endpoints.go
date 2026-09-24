package recipes

// Technological charts — JSON, iiko 6.0.
const (
	EndpointChartsGetAll       = "/api/v2/assemblyCharts/getAll"
	EndpointChartsGetAllUpdate = "/api/v2/assemblyCharts/getAllUpdate"
	EndpointChartsGetTree      = "/api/v2/assemblyCharts/getTree"
	EndpointChartsGetAssembled = "/api/v2/assemblyCharts/getAssembled"
	EndpointChartsGetPrepared  = "/api/v2/assemblyCharts/getPrepared"
	EndpointChartsByID         = "/api/v2/assemblyCharts/byId"
	EndpointChartsGetHistory   = "/api/v2/assemblyCharts/getHistory" // productId required, departmentId optional
	EndpointChartsSave         = "/api/v2/assemblyCharts/save"       // writes: changes what gets written off
	EndpointChartsDelete       = "/api/v2/assemblyCharts/delete"     // writes
)
