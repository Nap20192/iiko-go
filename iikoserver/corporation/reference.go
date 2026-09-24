package corporation

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Settings are the corporation-wide accounting settings. Requires B_ADM.
type Settings struct {
	VatAccounting string `json:"vatAccounting"` // VAT_INCLUDED_IN_PRICE | VAT_NOT_INCLUDED_IN_PRICE
}

// Settings reads the corporation settings.
//
// The docs print this path both with and without /v2/ on the same page, so the
// /v2/ spelling is tried first and only a 404 falls back: a 403 is a permission
// problem and retrying it as a wrong path would hide that.
func (s *Service) Settings(ctx context.Context) (*Settings, error) {
	var out Settings
	_, err := s.rest.GetV2(ctx, EndpointSettings, nil, &out)
	if rest.IsNotFound(err) {
		_, err = s.rest.GetV2(ctx, EndpointSettingsV1, nil, &out)
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ReferenceEntity is one row of /v2/entities/list.
//
// The docs warn this is for display names only: it carries no department
// binding and no validity periods, so it must not drive business decisions.
type ReferenceEntity struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	RootType string `json:"rootType"`
	Deleted  bool   `json:"deleted"`
}

// ListReferenceEntities reads reference books by root type.
//
// includeDeleted defaults to TRUE here, the opposite of nearly every other
// endpoint, so it is always sent explicitly.
func (s *Service) ListReferenceEntities(ctx context.Context, rootTypes []string, includeDeleted bool, revisionFrom int) ([]ReferenceEntity, error) {
	if len(rootTypes) == 0 {
		return nil, fmt.Errorf("at least one rootType is required: an unfiltered list spans every reference book")
	}
	q := url.Values{
		"includeDeleted": {rest.BoolStr(includeDeleted)},
		"revisionFrom":   {fmt.Sprint(revisionFrom)},
	}
	for _, rt := range rootTypes {
		q.Add("rootType", rt)
	}
	var out []ReferenceEntity
	if _, err := s.rest.GetV2(ctx, EndpointEntitiesList, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// EntityIDs lists just the ids of one reference book, iiko 9.1.
func (s *Service) EntityIDs(ctx context.Context, entityType string, includeDeleted bool, revisionFrom int) ([]string, error) {
	if strings.TrimSpace(entityType) == "" {
		return nil, fmt.Errorf("entityType is required")
	}
	path := strings.Replace(EndpointEntityIDs, "{entityType}", url.PathEscape(entityType), 1)
	q := url.Values{
		"includeDeleted": {rest.BoolStr(includeDeleted)},
		"revisionFrom":   {fmt.Sprint(revisionFrom)},
	}
	var out []string
	if _, err := s.rest.GetV2(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SearchDepartments matches on code. The parameter is a REGEX: a plain string
// matches case-sensitively anywhere in the value, not as a whole-field equality.
func (s *Service) SearchDepartments(ctx context.Context, codeRegex string) ([]Entity, error) {
	return s.searchCorporate(ctx, EndpointDepartmentsSearch, url.Values{"code": {codeRegex}})
}

// SearchStores matches stores on code.
//
// Store codes are optional in iiko, so a base where nobody filled them answers
// an empty list rather than an error.
func (s *Service) SearchStores(ctx context.Context, codeRegex string) ([]Entity, error) {
	return s.searchCorporate(ctx, EndpointStoresSearch, url.Values{"code": {codeRegex}})
}

// SearchGroups matches till groups on name, optionally within one department.
func (s *Service) SearchGroups(ctx context.Context, nameRegex, departmentID string) ([]Entity, error) {
	q := url.Values{"name": {nameRegex}}
	if departmentID != "" {
		q.Set("departmentId", departmentID)
	}
	return s.searchCorporate(ctx, EndpointGroupsSearch, q)
}

// SearchTerminals matches terminals on the documented regex fields. Front
// terminals are the ones with anonymous=false.
func (s *Service) SearchTerminals(ctx context.Context, fields map[string]string) ([]Entity, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("at least one search field is required")
	}
	q := url.Values{}
	for k, v := range fields {
		q.Set(k, v)
	}
	return s.searchCorporate(ctx, EndpointTerminalsSearch, q)
}

func (s *Service) searchCorporate(ctx context.Context, path string, q url.Values) ([]Entity, error) {
	var raw corporateItems
	if err := s.rest.GetXML(ctx, path, q, &raw); err != nil {
		return nil, err
	}
	out := make([]Entity, 0, len(raw.Items))
	for _, it := range raw.Items {
		out = append(out, Entity{ID: it.ID, ParentID: it.ParentID, Code: it.Code, Name: it.Name, Type: it.Type})
	}
	return out, nil
}

// ReplicationStatus is one department's replication state.
type ReplicationStatus struct {
	DepartmentID string `json:"departmentId"`
	Status       string `json:"status"`
}

// ReplicationStatuses lists replication state for every department.
//
// Chain only: on an RMS this errors outright, so gate on ServerType first.
func (s *Service) ReplicationStatuses(ctx context.Context) ([]ReplicationStatus, error) {
	var out []ReplicationStatus
	if _, err := s.rest.GetV2(ctx, EndpointReplicationStatuses, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ReplicationStatus reads one department's replication state. Chain only.
func (s *Service) ReplicationStatus(ctx context.Context, departmentID string) (*ReplicationStatus, error) {
	if strings.TrimSpace(departmentID) == "" {
		return nil, fmt.Errorf("departmentId is required")
	}
	path := strings.Replace(EndpointReplicationStatus, "{departmentId}", url.PathEscape(departmentID), 1)
	var out ReplicationStatus
	if _, err := s.rest.GetV2(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
