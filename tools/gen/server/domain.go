package main

import "regexp"

// domainRules maps an endpoint path to the domain that owns it, in order.
var domainRules = []struct {
	re  *regexp.Regexp
	dom string
}{
	{regexp.MustCompile(`^/resto/api/(v2/)?entities/accounts/list$`), "corporation"},
	{regexp.MustCompile(`^/resto/api/corporation/`), "corporation"},
	{regexp.MustCompile(`^/resto/api/replication/`), "corporation"},
	{regexp.MustCompile(`^/resto/api/v2/entities/list$`), "corporation"},
	{regexp.MustCompile(`^/resto/api/v2/entities/\{entityType\}/ids$`), "corporation"},
	{regexp.MustCompile(`^/resto/api/v2/documents/menuChange`), "pricing"},
	{regexp.MustCompile(`^/resto/api/v2/price`), "pricing"},
	{regexp.MustCompile(`^/resto/api/v2/entities/(priceCategories|periodSchedules)`), "pricing"},
	{regexp.MustCompile(`^/resto/api/v2/entities/products`), "nomenclature"},
	{regexp.MustCompile(`^/resto/api/v2/entities/productScales`), "nomenclature"},
	{regexp.MustCompile(`^/resto/api/v2/entities/quickLabels`), "nomenclature"},
	{regexp.MustCompile(`^/resto/api/v2/images/`), "nomenclature"},
	{regexp.MustCompile(`^/resto/api/products`), "nomenclature"},
	{regexp.MustCompile(`^/resto/api/v2/assemblyCharts/`), "recipes"},
	{regexp.MustCompile(`^/resto/api/documents/`), "documents"},
	{regexp.MustCompile(`^/resto/api/v2/documents/`), "documents"},
	{regexp.MustCompile(`^/resto/api/(v2/)?reports/`), "reports"},
	{regexp.MustCompile(`^/resto/api/employees`), "staff"},
	{regexp.MustCompile(`^/resto/api/v2/(cashshifts|payInOuts|payrolls)`), "cashshifts"},
	{regexp.MustCompile(`^/resto/api/v2/entities/payInOutTypes`), "cashshifts"},
	{regexp.MustCompile(`^/resto/api/events`), "events"},
	{regexp.MustCompile(`^/resto/api/edi/`), "edi"},
	{regexp.MustCompile(`^/resto/api/suppliers`), "suppliers"},
	{regexp.MustCompile(`^/resto/api/(auth|logout|licence)`), "rest"},
}

func domainOf(path string) string {
	for _, r := range domainRules {
		if r.re.MatchString(path) {
			return r.dom
		}
	}
	return ""
}

// overrides settle DTOs the path rules cannot: shared across domains, or
// reachable only as another DTO's field. Each records why.
var overrides = map[string]string{
	// suppliers ARE employees with supplier=true, so staff owns the shape
	"EmployeeDto": "staff",
	// three of four owners are nomenclature; the supplier pricelist reuses it
	"ContainerDto__nomenclature": "nomenclature",
	// base-type vocabulary from kody-bazovykh-tipov; report columns use it most
	"AccountGroup": "reports", "AccountType": "reports", "AlcoholClassType": "reports",
	"BanquetFlag": "reports", "CashFlowCategoryType": "reports", "CounteragentType": "reports",
	"DeliveryFlag": "reports", "DepartmentType": "reports", "DishRemovalType": "reports",
	"DocumentType": "reports", "GoodsType": "reports", "IsCashFlowAccount": "reports",
	"OperationType": "reports", "OrderDeletedFlag": "reports", "PaymentFiscalFlag": "reports",
	"PaymentGroup": "reports", "ProductGroupType": "reports", "ProductType": "reports",
	"StoreOrAccount": "reports", "TransactionSide": "reports", "TransactionType": "reports",
	"OlapFieldsUnknown": "reports", "OlapFilter": "reports",
}

// skipped DTOs are deliberately not generated, with the reason.
var skipped = map[string]string{
	"ApiError":                "the transport owns the error model (rest.Error)",
	"ErrorDto":                "the v2 envelope owns it (rest, envelope.go)",
	"OlapBizRequest":          "iiko.biz, a different product",
	"OlapBizByPresetRequest":  "iiko.biz, a different product",
	"OlapBizByPresetResponse": "iiko.biz, a different product",
	"OlapBizPresetsResponse":  "iiko.biz, a different product",
}

// duplicated DTOs are emitted into every domain that references them: they are
// small leaf shapes used across domains, and a shared package is not in the
// layout. Losing them to json.RawMessage would hide the error code from callers.
var duplicated = map[string]bool{"ErrorDto": true}
