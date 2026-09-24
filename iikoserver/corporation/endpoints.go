package corporation

// Corporation structure and reference books.
const (
	EndpointServerType     = "/api/replication/serverType" // CHAIN | REPLICATED_RMS | STANDALONE_RMS
	EndpointDepartments    = "/api/corporation/departments"
	EndpointStores         = "/api/corporation/stores"
	EndpointGroups         = "/api/corporation/groups"
	EndpointTerminals      = "/api/corporation/terminals"
	EndpointAccountsList   = "/api/v2/entities/accounts/list" // JSON, bare array
	EndpointAccountsListV1 = "/api/entities/accounts/list"    // same, header spelling

	// Search variants. `code` and `name` are REGEXes, not substrings: a plain
	// string matches case-sensitively anywhere in the value.
	EndpointDepartmentsSearch = "/api/corporation/departments/search"
	EndpointStoresSearch      = "/api/corporation/stores/search" // verb inferred: no CSS block, example URL only
	EndpointGroupsSearch      = "/api/corporation/groups/search"
	EndpointTerminalsSearch   = "/api/corporation/terminals/search"

	// Settings is JSON on a v1-looking path, and the docs print it both with and
	// without /v2/ on the same page. Both are tried, /v2/ first.
	EndpointSettings   = "/api/v2/corporation/settings" // JSON, right B_ADM
	EndpointSettingsV1 = "/api/corporation/settings"    // same, block-header spelling

	// Reference books.
	EndpointEntitiesList = "/api/v2/entities/list"             // JSON; includeDeleted defaults TRUE here
	EndpointEntityIDs    = "/api/v2/entities/{entityType}/ids" // JSON, iiko 9.1

	// Replication — Chain only; these error outright on an RMS.
	EndpointReplicationStatuses = "/api/replication/statuses"
	EndpointReplicationStatus   = "/api/replication/byDepartmentId/{departmentId}/status"
)
