package nomenclature

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// SectionDepartmentWide is the literal value sectionId takes for the menu that
// covers a whole department rather than one restaurant section.
const SectionDepartmentWide = "null"

// QuickMenuFilter narrows a quick-menu list. The id parameter is singular here,
// unlike every other list in this domain; "ids" is silently ignored upstream.
type QuickMenuFilter struct {
	IncludeDeleted bool
	IDs            []string
	DepartmentIDs  []string
	SectionIDs     []string // SectionDepartmentWide selects the department-wide menu
	RevisionFrom   *int64   // no documented default, so it is sent only when set
}

func (f QuickMenuFilter) values() url.Values {
	v := url.Values{"includeDeleted": {rest.BoolStr(f.IncludeDeleted)}}
	addAll(v, "id", f.IDs)
	addAll(v, "departmentId", f.DepartmentIDs)
	addAll(v, "sectionId", f.SectionIDs)
	if f.RevisionFrom != nil {
		v.Set("revisionFrom", strconv.FormatInt(*f.RevisionFrom, 10))
	}
	return v
}

// ListQuickMenus reads the quick menus of a department over GET, both the
// department-wide one and any per-section ones. Requires the B_QMENU right.
func (s *Service) ListQuickMenus(ctx context.Context, f QuickMenuFilter) ([]QuickMenuDto, error) {
	var out []QuickMenuDto
	if _, err := s.rest.GetV2(ctx, EndpointQuickLabelsList, f.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// QueryQuickMenus reads them over the POST twin of the same path.
func (s *Service) QueryQuickMenus(ctx context.Context, f QuickMenuFilter) ([]QuickMenuDto, error) {
	var out []QuickMenuDto
	if err := s.postForm(ctx, EndpointQuickLabelsList, f.values(), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// quickLabelWrite exists for day alone. day is nullable upstream and 0 means
// Monday, so neither int nor omitempty can carry it; the other fields are
// written out with it because encoding/json cannot override one field of a type
// that owns a MarshalJSON. page, x and y now match the generated shape exactly.
type quickLabelWrite struct {
	Day      *int   `json:"day"`
	Page     int    `json:"page"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	EntityID string `json:"entityId"`
}

type quickMenuWrite struct {
	ID               string            `json:"id,omitempty"`
	DependsOnWeekDay bool              `json:"dependsOnWeekDay"`
	DepartmentID     string            `json:"departmentId,omitempty"`
	SectionID        string            `json:"sectionId,omitempty"`
	PageNames        []string          `json:"pageNames,omitempty"`
	Labels           []quickLabelWrite `json:"labels"`
}

// writeBody converts a menu for the wire, checking the documented grid bounds
// and deciding each label's day from the menu's own dependsOnWeekDay.
func writeBody(m QuickMenuDto) (quickMenuWrite, error) {
	out := quickMenuWrite{
		ID: m.ID, DependsOnWeekDay: m.DependsOnWeekDay, DepartmentID: m.DepartmentID,
		SectionID: m.SectionID, PageNames: m.PageNames,
		Labels: make([]quickLabelWrite, 0, len(m.Labels)),
	}
	for i, l := range m.Labels {
		if l.Page < 0 || l.Page > 2 {
			return out, fmt.Errorf("label %d: page %d is out of range, the menu has 3 pages (0..2)", i, l.Page)
		}
		if l.X < 0 || l.X > 2 {
			return out, fmt.Errorf("label %d: x %d is out of range, the grid has 3 columns (0..2)", i, l.X)
		}
		if l.Y < 0 || l.Y > 7 {
			return out, fmt.Errorf("label %d: y %d is out of range, the grid has 8 rows (0..7)", i, l.Y)
		}
		w := quickLabelWrite{Page: l.Page, X: l.X, Y: l.Y, EntityID: l.EntityID}
		// day is null on every label of a menu that is not day-of-week specific;
		// sending 0 there would pin the whole menu to Mondays.
		if m.DependsOnWeekDay {
			if _, err := QuickLabelDay(l); err != nil {
				return out, fmt.Errorf("label %d: %w", i, err)
			}
			day := l.Day
			w.Day = &day
		}
		out.Labels = append(out.Labels, w)
	}
	return out, nil
}

// SaveQuickMenu creates a quick menu. It changes production data: this is what
// a cashier sees on the till.
func (s *Service) SaveQuickMenu(ctx context.Context, m QuickMenuDto) (*QuickMenuDto, error) {
	body, err := writeBody(m)
	if err != nil {
		return nil, err
	}
	body.ID = ""
	return s.quickMenuWrite(ctx, EndpointQuickLabelsSave, body)
}

// UpdateQuickMenu edits an existing quick menu.
func (s *Service) UpdateQuickMenu(ctx context.Context, m QuickMenuDto) (*QuickMenuDto, error) {
	if strings.TrimSpace(m.ID) == "" {
		return nil, fmt.Errorf("id is required on update: without it the request addresses no menu")
	}
	body, err := writeBody(m)
	if err != nil {
		return nil, err
	}
	return s.quickMenuWrite(ctx, EndpointQuickLabelsUpdate, body)
}

// DeleteQuickMenu soft-deletes a quick menu. The body is a bare id.
func (s *Service) DeleteQuickMenu(ctx context.Context, id string) (*QuickMenuDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id is required")
	}
	return s.quickMenuWrite(ctx, EndpointQuickLabelsDelete, categoryRef{ID: id})
}

func (s *Service) quickMenuWrite(ctx context.Context, path string, body any) (*QuickMenuDto, error) {
	var out QuickMenuDto
	if _, err := s.rest.PostV2(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// QuickLabelDay reads a label's weekday. Quick menus number days 0=Monday..
// 6=Sunday; period schedules in pkg/iikoserver/pricing number them 1..7. Same
// API, two conventions, and neither errors when read with the other.
func QuickLabelDay(l QuickLabelDto) (time.Weekday, error) {
	if l.Day < 0 || l.Day > 6 {
		return 0, fmt.Errorf("day %d is out of range: quick menus use 0=Monday..6=Sunday", l.Day)
	}
	return time.Weekday((l.Day + 1) % 7), nil
}

// imageUpload is the documented save body: the field table names the Base64
// payload "id", and the id in the response is the saved image's real UUID.
type imageUpload struct {
	ID string `json:"id"`
}

// LoadImage reads an image as Base64.
func (s *Service) LoadImage(ctx context.Context, imageID string) (*ImageDto, error) {
	if strings.TrimSpace(imageID) == "" {
		return nil, fmt.Errorf("imageId is required")
	}
	var out ImageDto
	if _, err := s.rest.GetV2(ctx, EndpointImagesLoad, url.Values{"imageId": {imageID}}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SaveImage uploads a Base64 image and returns it with the id iiko assigned.
// The server caps the size at its saved-image-max-size-mb setting, 512 MB by
// default, which is not something a client can check.
func (s *Service) SaveImage(ctx context.Context, base64Data string) (*ImageDto, error) {
	if strings.TrimSpace(base64Data) == "" {
		return nil, fmt.Errorf("the Base64 payload is required")
	}
	var out ImageDto
	if _, err := s.rest.PostV2(ctx, EndpointImagesSave, nil, imageUpload{ID: base64Data}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteImages removes images and returns the ids actually deleted.
func (s *Service) DeleteImages(ctx context.Context, ids []string) ([]string, error) {
	if err := requireIDs(ids); err != nil {
		return nil, err
	}
	var out IdListDto
	if _, err := s.rest.PostV2(ctx, EndpointImagesDelete, nil, idList(ids), &out); err != nil {
		return nil, err
	}
	deleted := make([]string, len(out.Items))
	for i, item := range out.Items {
		deleted[i] = item.ID
	}
	return deleted, nil
}
