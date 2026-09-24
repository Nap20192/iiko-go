package suppliers

// Supplier paths. Suppliers are Users with supplier=true, so these answer the
// employee XML shape rather than a supplier-specific one.
const (
	EndpointSuppliers       = "/api/suppliers"
	EndpointSuppliersSearch = "/api/suppliers/search"
	// {code} is the supplier CODE, not a UUID, and this endpoint's `date` is
	// DD.MM.YYYY while its neighbours take ISO.
	EndpointSupplierPricelist = "/api/suppliers/{code}/pricelist"
)
