// Package menu covers external menus, stop lists, combos and the nomenclature catalogue.
package menu

import (
	"context"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// Service is the menu half of iikoCloud.
type Service struct{ c *rest.Client }

// New wires a service onto a transport; the facade is the usual caller.
func New(c *rest.Client) *Service { return &Service{c: c} }

// AddStopLists add items to out-of-stock list. (You should have extra rights to use this method).
func (s *Service) AddStopLists(ctx context.Context, req gen.AddProductsToStopListRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, addStopLists, req)
}

// CalculateCombo calculate combo price.
func (s *Service) CalculateCombo(ctx context.Context, req gen.CalculateComboPriceRequest) (*gen.CalculateComboPriceResponse, error) {
	return rest.Call[gen.CalculateComboPriceResponse](ctx, s.c, calculateCombo, req)
}

// CheckStopLists check items in out-of-stock list.
func (s *Service) CheckStopLists(ctx context.Context, req gen.CheckStopListRequest) (*gen.CheckStopListResponse, error) {
	return rest.Call[gen.CheckStopListResponse](ctx, s.c, checkStopLists, req)
}

// ClearStopLists clear out-of-stock list. (You should have extra rights to use this method).
func (s *Service) ClearStopLists(ctx context.Context, req gen.ClearStopListRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, clearStopLists, req)
}

// CreateNomenclatureAssemblyChart create an assembly chart (v2).
func (s *Service) CreateNomenclatureAssemblyChart(ctx context.Context, req gen.CreateRequest) error {
	return s.c.Post(ctx, createNomenclatureAssemblyChart, req, nil)
}

// CreateNomenclatureCategory create a product category.
func (s *Service) CreateNomenclatureCategory(ctx context.Context, req gen.NomenclatureCategoryCreateRequest) error {
	return s.c.Post(ctx, createNomenclatureCategory, req, nil)
}

// CreateNomenclatureGroup create a nomenclature group (v2).
func (s *Service) CreateNomenclatureGroup(ctx context.Context, req gen.NomenclatureGroupCreateRequest) error {
	return s.c.Post(ctx, createNomenclatureGroup, req, nil)
}

// CreateNomenclatureProduct create a product (v2).
func (s *Service) CreateNomenclatureProduct(ctx context.Context, req gen.NomenclatureProductCreateRequest) error {
	return s.c.Post(ctx, createNomenclatureProduct, req, nil)
}

// CreateNomenclatureProductScale create a product size scale.
func (s *Service) CreateNomenclatureProductScale(ctx context.Context, req gen.ProductSizeCreateRequest) (*gen.ProductScaleResponse, error) {
	return rest.Call[gen.ProductScaleResponse](ctx, s.c, createNomenclatureProductScale, req)
}

// DeleteNomenclatureAssemblyChart delete an assembly chart (v2).
func (s *Service) DeleteNomenclatureAssemblyChart(ctx context.Context, req gen.DeleteRequest) (*gen.AssemblyChartCommonResponse, error) {
	return rest.Call[gen.AssemblyChartCommonResponse](ctx, s.c, deleteNomenclatureAssemblyChart, req)
}

// DeleteNomenclatureCategory delete a product category.
func (s *Service) DeleteNomenclatureCategory(ctx context.Context, req gen.NomenclatureCategoryDeleteRequest) (*gen.NomenclatureCategoryWriteResponse, error) {
	return rest.Call[gen.NomenclatureCategoryWriteResponse](ctx, s.c, deleteNomenclatureCategory, req)
}

// DeleteNomenclatureGroup delete nomenclature groups (v2).
func (s *Service) DeleteNomenclatureGroup(ctx context.Context, req gen.NomenclatureGroupDeleteRequest) (*gen.NomenclatureGroupDeleteResponse, error) {
	return rest.Call[gen.NomenclatureGroupDeleteResponse](ctx, s.c, deleteNomenclatureGroup, req)
}

// DeleteNomenclatureProduct delete products (v2).
func (s *Service) DeleteNomenclatureProduct(ctx context.Context, req gen.NomenclatureProductDeleteRequest) (*gen.NomenclatureProductDeleteResponse, error) {
	return rest.Call[gen.NomenclatureProductDeleteResponse](ctx, s.c, deleteNomenclatureProduct, req)
}

// DeleteNomenclatureProductScale delete a product size scale.
func (s *Service) DeleteNomenclatureProductScale(ctx context.Context, req gen.ProductSizeDeleteRequest) (*gen.ProductScaleResponse, error) {
	return rest.Call[gen.ProductScaleResponse](ctx, s.c, deleteNomenclatureProductScale, req)
}

// GetCombo get combos info.
func (s *Service) GetCombo(ctx context.Context, req gen.GetCombosInfoRequest) (*gen.GetCombosInfoResponse, error) {
	return rest.Call[gen.GetCombosInfoResponse](ctx, s.c, getCombo, req)
}

// GetMenu external menus with price categories.
func (s *Service) GetMenu(ctx context.Context) (*gen.MenusDataResponse, error) {
	return rest.Call[gen.MenusDataResponse](ctx, s.c, getMenu, nil)
}

// GetMenuByID retrieve external menu V3 by ID.
func (s *Service) GetMenuByID(ctx context.Context, req gen.MenuRequestV3) (*gen.MenuV3, error) {
	return rest.Call[gen.MenuV3](ctx, s.c, getMenuByID, req)
}

// GetNomenclatureAssemblyChart get an assembly chart by ID (v2).
func (s *Service) GetNomenclatureAssemblyChart(ctx context.Context, req gen.GetRequest) (*gen.AssemblyChartResponse, error) {
	return rest.Call[gen.AssemblyChartResponse](ctx, s.c, getNomenclatureAssemblyChart, req)
}

// GetNomenclatureProductScale get a product size scale by ID.
func (s *Service) GetNomenclatureProductScale(ctx context.Context, req gen.ProductSizeGetRequest) (*gen.ProductScaleResponse, error) {
	return rest.Call[gen.ProductScaleResponse](ctx, s.c, getNomenclatureProductScale, req)
}

// GetStopLists out-of-stock items.
func (s *Service) GetStopLists(ctx context.Context, req gen.StopListsRequest) (*gen.StopListsResponse, error) {
	return rest.Call[gen.StopListsResponse](ctx, s.c, getStopLists, req)
}

// ListNomenclatureAllergenGroup get a list of allergen groups.
func (s *Service) ListNomenclatureAllergenGroup(ctx context.Context, req gen.AllergenGroupListRequest) (*gen.AllergenGroupListResponse, error) {
	return rest.Call[gen.AllergenGroupListResponse](ctx, s.c, listNomenclatureAllergenGroup, req)
}

// ListNomenclatureAmountUnit get a list of amount units.
func (s *Service) ListNomenclatureAmountUnit(ctx context.Context, req gen.AmountUnitListRequest) (*gen.AmountUnitListResponse, error) {
	return rest.Call[gen.AmountUnitListResponse](ctx, s.c, listNomenclatureAmountUnit, req)
}

// ListNomenclatureAssemblyChart get a list of assembly charts by product (v2).
func (s *Service) ListNomenclatureAssemblyChart(ctx context.Context, req gen.ListRequest) (*gen.ListResponse, error) {
	return rest.Call[gen.ListResponse](ctx, s.c, listNomenclatureAssemblyChart, req)
}

// ListNomenclatureCategory get a list of product categories.
func (s *Service) ListNomenclatureCategory(ctx context.Context, req gen.NomenclatureCategoryListRequest) (*gen.NomenclatureCategoryListResponse, error) {
	return rest.Call[gen.NomenclatureCategoryListResponse](ctx, s.c, listNomenclatureCategory, req)
}

// ListNomenclatureContainer get a list of containers.
func (s *Service) ListNomenclatureContainer(ctx context.Context, req gen.ContainerListRequest) (*gen.ContainerListResponse, error) {
	return rest.Call[gen.ContainerListResponse](ctx, s.c, listNomenclatureContainer, req)
}

// ListNomenclatureCustomCategory get a list of custom categories.
func (s *Service) ListNomenclatureCustomCategory(ctx context.Context, req gen.CustomCategoryListRequest) (*gen.CustomCategoryListResponse, error) {
	return rest.Call[gen.CustomCategoryListResponse](ctx, s.c, listNomenclatureCustomCategory, req)
}

// ListNomenclatureGroup get a list of nomenclature groups (v2).
func (s *Service) ListNomenclatureGroup(ctx context.Context, req gen.NomenclatureGroupListRequest) (*gen.NomenclatureGroupListResponse, error) {
	return rest.Call[gen.NomenclatureGroupListResponse](ctx, s.c, listNomenclatureGroup, req)
}

// ListNomenclatureMenu get a list of menu sections.
func (s *Service) ListNomenclatureMenu(ctx context.Context, req gen.MenuSectionListRequest) (*gen.MenuSectionListResponse, error) {
	return rest.Call[gen.MenuSectionListResponse](ctx, s.c, listNomenclatureMenu, req)
}

// ListNomenclatureModifierSchema get a list of modifier schemas.
func (s *Service) ListNomenclatureModifierSchema(ctx context.Context, req gen.ModifierSchemaListRequest) (*gen.ModifierSchemaListResponse, error) {
	return rest.Call[gen.ModifierSchemaListResponse](ctx, s.c, listNomenclatureModifierSchema, req)
}

// ListNomenclaturePlaceType get a list of preparation place types.
func (s *Service) ListNomenclaturePlaceType(ctx context.Context, req gen.PlaceTypeListRequest) (*gen.PlaceTypeListResponse, error) {
	return rest.Call[gen.PlaceTypeListResponse](ctx, s.c, listNomenclaturePlaceType, req)
}

// ListNomenclatureProducer get a list of producers.
func (s *Service) ListNomenclatureProducer(ctx context.Context, req gen.ProducerListRequest) (*gen.ProducerListResponse, error) {
	return rest.Call[gen.ProducerListResponse](ctx, s.c, listNomenclatureProducer, req)
}

// ListNomenclatureProduct get a list of products (v2).
func (s *Service) ListNomenclatureProduct(ctx context.Context, req gen.NomenclatureProductListRequest) (*gen.NomenclatureProductListResponse, error) {
	return rest.Call[gen.NomenclatureProductListResponse](ctx, s.c, listNomenclatureProduct, req)
}

// ListNomenclatureProductScale get a list of product size scales.
func (s *Service) ListNomenclatureProductScale(ctx context.Context, req gen.ProductSizeListRequest) (*gen.ProductScaleListResponse, error) {
	return rest.Call[gen.ProductScaleListResponse](ctx, s.c, listNomenclatureProductScale, req)
}

// ListNomenclatureProductSize get a list of product sizes.
func (s *Service) ListNomenclatureProductSize(ctx context.Context, req gen.ProductSizeListRequest) (*gen.ProductSizeListResponse, error) {
	return rest.Call[gen.ProductSizeListResponse](ctx, s.c, listNomenclatureProductSize, req)
}

// ListNomenclatureProductTag get a list of product tags.
func (s *Service) ListNomenclatureProductTag(ctx context.Context, req gen.ProductTagListRequest) (*gen.ProductTagListResponse, error) {
	return rest.Call[gen.ProductTagListResponse](ctx, s.c, listNomenclatureProductTag, req)
}

// RemoveStopLists remove items from out-of-stock list. (You should have extra rights to use this method).
func (s *Service) RemoveStopLists(ctx context.Context, req gen.RemoveProductsFromStopListRequest) (*gen.CorrelationIDResponse, error) {
	return rest.Call[gen.CorrelationIDResponse](ctx, s.c, removeStopLists, req)
}

// RestoreNomenclatureCategory restore a product category.
func (s *Service) RestoreNomenclatureCategory(ctx context.Context, req gen.NomenclatureCategoryRestoreRequest) (*gen.NomenclatureCategoryWriteResponse, error) {
	return rest.Call[gen.NomenclatureCategoryWriteResponse](ctx, s.c, restoreNomenclatureCategory, req)
}

// RestoreNomenclatureGroup restore nomenclature groups (v2).
func (s *Service) RestoreNomenclatureGroup(ctx context.Context, req gen.NomenclatureGroupUndeleteRequest) (*gen.NomenclatureGroupUndeleteResponse, error) {
	return rest.Call[gen.NomenclatureGroupUndeleteResponse](ctx, s.c, restoreNomenclatureGroup, req)
}

// RestoreNomenclatureProduct restore products (v2).
func (s *Service) RestoreNomenclatureProduct(ctx context.Context, req gen.NomenclatureProductUndeleteRequest) (*gen.NomenclatureProductUndeleteResponse, error) {
	return rest.Call[gen.NomenclatureProductUndeleteResponse](ctx, s.c, restoreNomenclatureProduct, req)
}

// UpdateBarcodesNomenclatureProduct update product barcodes (v2).
func (s *Service) UpdateBarcodesNomenclatureProduct(ctx context.Context, req gen.NomenclatureProductUpdateBarcodesRequest) (*gen.NomenclatureProductUpdateBarcodesResponse, error) {
	return rest.Call[gen.NomenclatureProductUpdateBarcodesResponse](ctx, s.c, updateBarcodesNomenclatureProduct, req)
}

// UpdateNomenclatureAssemblyChart update an assembly chart (v2).
func (s *Service) UpdateNomenclatureAssemblyChart(ctx context.Context, req gen.InternalDocumentAssemblyChartV2SaveRequest) (*gen.AssemblyChartCommonResponse, error) {
	return rest.Call[gen.AssemblyChartCommonResponse](ctx, s.c, updateNomenclatureAssemblyChart, req)
}

// UpdateNomenclatureCategory update a product category.
func (s *Service) UpdateNomenclatureCategory(ctx context.Context, req gen.NomenclatureCategoryUpdateRequest) (*gen.NomenclatureCategoryWriteResponse, error) {
	return rest.Call[gen.NomenclatureCategoryWriteResponse](ctx, s.c, updateNomenclatureCategory, req)
}

// UpdateNomenclatureGroup update a nomenclature group (v2).
func (s *Service) UpdateNomenclatureGroup(ctx context.Context, req gen.NomenclatureGroupUpdateRequest) (*gen.NomenclatureGroupUpdateResponse, error) {
	return rest.Call[gen.NomenclatureGroupUpdateResponse](ctx, s.c, updateNomenclatureGroup, req)
}

// UpdateNomenclatureProduct update a product (v2).
func (s *Service) UpdateNomenclatureProduct(ctx context.Context, req gen.NomenclatureProductUpdateRequest) (*gen.NomenclatureProductUpdateResponse, error) {
	return rest.Call[gen.NomenclatureProductUpdateResponse](ctx, s.c, updateNomenclatureProduct, req)
}

// UpdateNomenclatureProductScale update a product size scale.
func (s *Service) UpdateNomenclatureProductScale(ctx context.Context, req gen.ProductSizeUpdateRequest) (*gen.ProductScaleResponse, error) {
	return rest.Call[gen.ProductScaleResponse](ctx, s.c, updateNomenclatureProductScale, req)
}
