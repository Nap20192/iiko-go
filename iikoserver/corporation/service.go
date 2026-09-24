package corporation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// corporationEndpoint maps a reference-book kind to its path constant; "" for a
// kind this domain does not serve, so a new kind fails loudly.
func corporationEndpoint(kind EntityKind) string {
	switch kind {
	case KindDepartments:
		return EndpointDepartments
	case KindStores:
		return EndpointStores
	case KindGroups:
		return EndpointGroups
	case KindTerminals:
		return EndpointTerminals
	default:
		return ""
	}
}

// Service is the corporation structure and reference books half of the iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// EntityKind selects which reference book list_entities returns. One tool with
// an enum beats eight near-identical tools: the whole surface costs one schema.
type EntityKind string

const (
	KindDepartments EntityKind = "departments"
	KindStores      EntityKind = "stores"
	KindGroups      EntityKind = "groups"
	KindTerminals   EntityKind = "terminals"
	KindAccounts    EntityKind = "accounts"
	KindSuppliers   EntityKind = "suppliers"
)

func EntityKinds() []string {
	return []string{"departments", "stores", "groups", "terminals", "accounts"}
}

// Entity is the normalized shape across both response worlds. resto/api returns
// corporateItemDto (XML) for structure, an employee-shaped XML doc for
// suppliers, and JSON for accounts — callers should not care.
type Entity struct {
	ID       string `json:"id"`
	ParentID string `json:"parentId,omitempty"`
	Code     string `json:"code,omitempty"`
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Deleted  bool   `json:"deleted,omitempty"`
}

type corporateItems struct {
	Items []struct {
		ID       string `xml:"id"`
		ParentID string `xml:"parentId"`
		Code     string `xml:"code"`
		Name     string `xml:"name"`
		Type     string `xml:"type"`
	} `xml:"corporateItemDto"`
}

type account struct {
	ID       string `json:"id"`
	ParentID string `json:"accountParentId"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Deleted  bool   `json:"deleted"`
}

// ListEntities fetches one reference book.
//
// includeDeleted defaults are inconsistent across iiko endpoints (true on
// /v2/entities/*, false nearly everywhere else), so we always send it.
func (s *Service) ListEntities(ctx context.Context, kind EntityKind, includeDeleted bool) ([]Entity, error) {
	switch kind {
	case KindDepartments, KindStores, KindGroups, KindTerminals:
		path := corporationEndpoint(kind)
		if path == "" {
			return nil, fmt.Errorf("no endpoint mapped for corporation kind %q", kind)
		}
		var raw corporateItems
		if err := s.rest.GetXML(ctx, path, nil, &raw); err != nil {
			return nil, err
		}
		out := make([]Entity, 0, len(raw.Items))
		for _, it := range raw.Items {
			out = append(out, Entity{ID: it.ID, ParentID: it.ParentID, Code: it.Code, Name: it.Name, Type: it.Type})
		}
		return out, nil

	case KindAccounts:
		q := url.Values{"includeDeleted": {rest.BoolStr(includeDeleted)}}
		var raw []account
		// Two spellings are documented and we cannot tell which a given build
		// honours, so try the example's form then the header's.
		_, err := s.rest.GetV2(ctx, EndpointAccountsList, q, &raw)
		if rest.IsNotFound(err) {
			_, err = s.rest.GetV2(ctx, EndpointAccountsListV1, q, &raw)
		}
		if err != nil {
			return nil, err
		}
		out := make([]Entity, 0, len(raw))
		for _, a := range raw {
			if a.Deleted && !includeDeleted {
				continue
			}
			// Written out rather than converted: account is the wire shape and
			// Entity the domain one, and they are expected to diverge.
			out = append(out, Entity{ //nolint:staticcheck // S1016: see above
				ID: a.ID, ParentID: a.ParentID, Code: a.Code,
				Name: a.Name, Type: a.Type, Deleted: a.Deleted,
			})
		}
		return out, nil

	case KindSuppliers:
		return nil, fmt.Errorf("suppliers moved to their own domain: use Client.Suppliers.List — they are Users with supplier=true and share only a response format with the corporation structure")
	}
	return nil, fmt.Errorf("unknown entity kind %q; valid kinds: %s", kind, strings.Join(EntityKinds(), ", "))
}

// ServerInfo probes what this server can do before tools assume it.
// Chain-only endpoints error outright on an RMS, so serverType is the gate.
type ServerInfo struct {
	ServerType string `json:"serverType"` // CHAIN | REPLICATED_RMS | STANDALONE_RMS
}

func (s *Service) ServerType(ctx context.Context) (string, error) {
	data, err := s.rest.Do(ctx, http.MethodGet, EndpointServerType, nil, nil, "")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.Trim(string(data), `"`)), nil
}
