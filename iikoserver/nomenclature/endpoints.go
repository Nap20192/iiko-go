package nomenclature

// The nomenclature catalogue: products, the group and category trees they hang
// in, size scales, quick menus and images.
//
// Every /v2/ list has a GET and a POST form on one path. They are separate
// catalog entries but one URL, so one constant serves both and the verb is the
// method's business. /entities/products/{productId}/productScale goes further:
// GET, POST and DELETE on that URL are three different operations, which the
// docs never say — only their CSS-tagged request blocks reveal it.
//
// Not here on purpose: /v2/entities/products/sizes/list. It is in the catalog
// only because a third-party client calls it — undocumented, verb unconfirmed,
// no response shape. Use ProductScales or /v2/entities/list with
// rootType=ProductSize instead.
const (
	// Legacy v1 XML. Pulls supplier goods too, and those cannot be deleted.
	EndpointProductsXML       = "/api/products"        // XML
	EndpointProductsSearchXML = "/api/products/search" // XML, every filter is a regular expression

	// Products, v2 JSON, right B_EN.
	EndpointProductsList    = "/api/v2/entities/products/list"    // JSON, bare array; GET and POST
	EndpointProductsSave    = "/api/v2/entities/products/save"    // JSON, v2-wrapped
	EndpointProductsUpdate  = "/api/v2/entities/products/update"  // JSON, v2-wrapped
	EndpointProductsDelete  = "/api/v2/entities/products/delete"  // JSON, v2-wrapped
	EndpointProductsRestore = "/api/v2/entities/products/restore" // JSON, v2-wrapped

	// Nomenclature groups.
	EndpointGroupList    = "/api/v2/entities/products/group/list"    // JSON, bare array; GET and POST
	EndpointGroupSave    = "/api/v2/entities/products/group/save"    // JSON, v2-wrapped
	EndpointGroupUpdate  = "/api/v2/entities/products/group/update"  // JSON, v2-wrapped
	EndpointGroupDelete  = "/api/v2/entities/products/group/delete"  // JSON, v2-wrapped; products and groups together
	EndpointGroupRestore = "/api/v2/entities/products/group/restore" // JSON, v2-wrapped; products and groups together

	// User-defined product categories, not tax or accounting categories.
	EndpointCategoryList    = "/api/v2/entities/products/category/list"    // JSON, bare array; GET and POST
	EndpointCategorySave    = "/api/v2/entities/products/category/save"    // JSON, v2-wrapped
	EndpointCategoryUpdate  = "/api/v2/entities/products/category/update"  // JSON, v2-wrapped
	EndpointCategoryDelete  = "/api/v2/entities/products/category/delete"  // JSON, v2-wrapped
	EndpointCategoryRestore = "/api/v2/entities/products/category/restore" // JSON, v2-wrapped

	// Size scales: the definitions.
	EndpointProductScales       = "/api/v2/entities/productScales"         // JSON, bare array; GET and POST
	EndpointProductScaleByID    = "/api/v2/entities/productScales/"        // JSON, bare object; + {productScaleId}
	EndpointProductScaleSave    = "/api/v2/entities/productScales/save"    // JSON, v2-wrapped
	EndpointProductScaleUpdate  = "/api/v2/entities/productScales/update"  // JSON, v2-wrapped
	EndpointProductScaleDelete  = "/api/v2/entities/productScales/delete"  // JSON, v2-wrapped
	EndpointProductScaleRestore = "/api/v2/entities/productScales/restore" // JSON, v2-wrapped

	// Size scales: which scale a product is bound to, and with what overrides.
	EndpointProductScaleBinding = "/api/v2/entities/products/%s/productScale" // JSON; GET reads, POST binds, DELETE unbinds
	EndpointProductScalesBatch  = "/api/v2/entities/products/productScales"   // JSON, map keyed by productId; GET and POST

	// Quick menus, right B_QMENU.
	EndpointQuickLabelsList   = "/api/v2/entities/quickLabels/list"   // JSON, bare array; GET and POST
	EndpointQuickLabelsSave   = "/api/v2/entities/quickLabels/save"   // JSON, bare object
	EndpointQuickLabelsUpdate = "/api/v2/entities/quickLabels/update" // JSON, bare object
	EndpointQuickLabelsDelete = "/api/v2/entities/quickLabels/delete" // JSON, bare object

	// Images, Base64 in and out.
	EndpointImagesLoad   = "/api/v2/images/load"   // JSON, bare object
	EndpointImagesSave   = "/api/v2/images/save"   // JSON, v2-wrapped
	EndpointImagesDelete = "/api/v2/images/delete" // JSON, v2-wrapped
)
