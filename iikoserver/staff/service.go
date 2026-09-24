package staff

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
)

// Service is the employees, schedules, attendances and waiter-teams half of the
// iikoServer API.
type Service struct{ rest *rest.Client }

// New binds this domain to the shared transport: one client, one licence seat.
func New(c *rest.Client) *Service { return &Service{rest: c} }

// ShiftFilter scopes a schedule or attendance read.
//
// From and To are both INCLUSIVE here. The docs call this out as an exception to
// the rest of the API, so asking for one day means From == To.
//
// A department may be named by code or by id but not both: iiko exposes a
// separate endpoint for each and none for the pair.
type ShiftFilter struct {
	From, To           time.Time
	EmployeeID         string
	DepartmentCode     string
	DepartmentID       string
	WithPaymentDetails bool // true fails outright if the employee has an unclosed earlier attendance
	RevisionFrom       int  // 0 means "unset"; the wire default is -1
}

func (f ShiftFilter) query() url.Values {
	rev := f.RevisionFrom
	if rev == 0 {
		rev = -1
	}
	return url.Values{
		"from":               {f.From.Format(rest.QueryV2)},
		"to":                 {f.To.Format(rest.QueryV2)},
		"withPaymentDetails": {rest.BoolStr(f.WithPaymentDetails)},
		"revisionFrom":       {fmt.Sprint(rev)},
	}
}

// shiftPath picks the read shape matching the scope. base is the unscoped path;
// byEmployee, byDeptCode and byDeptID are its scoped siblings.
func shiftPath(f ShiftFilter, base, byEmployee, byDeptCode, byDeptID string) (string, error) {
	if f.DepartmentCode != "" && f.DepartmentID != "" {
		return "", fmt.Errorf("scope by department_code or department_id, not both: iiko has an endpoint for each and none for the pair")
	}
	switch {
	case f.DepartmentID != "" && f.EmployeeID != "":
		return byDeptID + url.PathEscape(f.DepartmentID) + "/byEmployee/" + url.PathEscape(f.EmployeeID) + "/", nil
	case f.DepartmentID != "":
		return byDeptID + url.PathEscape(f.DepartmentID) + "/", nil
	case f.DepartmentCode != "" && f.EmployeeID != "":
		return byDeptCode + url.PathEscape(f.DepartmentCode) + "/byEmployee/" + url.PathEscape(f.EmployeeID) + "/", nil
	case f.DepartmentCode != "":
		return byDeptCode + url.PathEscape(f.DepartmentCode) + "/", nil
	case f.EmployeeID != "":
		return byEmployee + url.PathEscape(f.EmployeeID) + "/", nil
	default:
		return base, nil
	}
}

// trimTemplate turns a path constant's {placeholder} tail into a prefix to
// append to, keeping the full template greppable in endpoints.go.
func trimTemplate(p string) string {
	if i := strings.Index(p, "{"); i >= 0 {
		return p[:i]
	}
	return p
}

type scheduleList struct {
	Items []ScheduleDto `xml:"schedule"`
}

// Schedules lists planned shifts overlapping [From, To], both inclusive.
func (s *Service) Schedules(ctx context.Context, f ShiftFilter) ([]ScheduleDto, error) {
	path, err := shiftPath(f, EndpointSchedule,
		trimTemplate(EndpointScheduleByEmployee),
		trimTemplate(EndpointScheduleByDepartmentCode),
		trimTemplate(EndpointScheduleByDepartmentID))
	if err != nil {
		return nil, err
	}
	var out scheduleList
	if err := s.rest.GetXML(ctx, path, f.query(), &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

type attendanceList struct {
	Items []AttendanceDto `xml:"attendance"`
}

// Attendances lists worked shifts overlapping [From, To], both inclusive.
func (s *Service) Attendances(ctx context.Context, f ShiftFilter) ([]AttendanceDto, error) {
	path, err := shiftPath(f, EndpointAttendance,
		trimTemplate(EndpointAttendanceByEmployee),
		trimTemplate(EndpointAttendanceByDepartmentCode),
		trimTemplate(EndpointAttendanceByDepartmentID))
	if err != nil {
		return nil, err
	}
	var out attendanceList
	if err := s.rest.GetXML(ctx, path, f.query(), &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// UpdateSchedule edits a shift and returns the stored one.
//
// The docs warn that updating MAY CHANGE the shift's id, so the returned value
// is the only safe identity to keep; the submitted id may already be stale.
func (s *Service) UpdateSchedule(ctx context.Context, sch ScheduleDto) (*ScheduleDto, error) {
	return s.writeSchedule(ctx, EndpointScheduleUpdate, sch)
}

// CreateSchedule adds a shift and returns the stored one, id included.
func (s *Service) CreateSchedule(ctx context.Context, sch ScheduleDto) (*ScheduleDto, error) {
	return s.writeSchedule(ctx, EndpointScheduleCreate, sch)
}

func (s *Service) writeSchedule(ctx context.Context, endpoint string, sch ScheduleDto) (*ScheduleDto, error) {
	if strings.TrimSpace(sch.EmployeeID) == "" {
		return nil, fmt.Errorf("employeeId is required on a shift")
	}
	var out scheduleList
	if err := s.rest.PostXML(ctx, endpoint, nil, sch, &out); err != nil {
		return nil, err
	}
	if len(out.Items) == 0 {
		return nil, fmt.Errorf("%s returned no shift; the id it assigned is unknown", endpoint)
	}
	return &out.Items[0], nil
}

// DeleteSchedule removes a shift.
func (s *Service) DeleteSchedule(ctx context.Context, id string) error {
	return s.deleteByID(ctx, trimTemplate(EndpointScheduleDelete), id, "schedule")
}

// CreateAttendance records a worked shift and returns the stored one.
func (s *Service) CreateAttendance(ctx context.Context, a AttendanceDto) (*AttendanceDto, error) {
	return s.writeAttendance(ctx, EndpointAttendanceCreate, a)
}

// UpdateAttendance edits a worked shift and returns the stored one. Like
// schedules, iiko MAY CHANGE the entry's id, so keep the returned identity.
func (s *Service) UpdateAttendance(ctx context.Context, a AttendanceDto) (*AttendanceDto, error) {
	return s.writeAttendance(ctx, EndpointAttendanceUpdate, a)
}

// DeleteAttendance removes a worked shift.
func (s *Service) DeleteAttendance(ctx context.Context, id string) error {
	return s.deleteByID(ctx, trimTemplate(EndpointAttendanceDelete), id, "attendance")
}

func (s *Service) writeAttendance(ctx context.Context, endpoint string, a AttendanceDto) (*AttendanceDto, error) {
	if strings.TrimSpace(a.EmployeeID) == "" {
		return nil, fmt.Errorf("employeeId is required on an attendance")
	}
	var out attendanceList
	if err := s.rest.PostXML(ctx, endpoint, nil, a, &out); err != nil {
		return nil, err
	}
	if len(out.Items) == 0 {
		return nil, fmt.Errorf("%s returned no attendance; the id it assigned is unknown", endpoint)
	}
	return &out.Items[0], nil
}

func (s *Service) deleteByID(ctx context.Context, prefix, id, what string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%s id is required", what)
	}
	_, err := s.rest.Do(ctx, http.MethodDelete, prefix+url.PathEscape(id), nil, nil, "")
	return err
}

type scheduleTypeList struct {
	Items []ScheduleTypeDto `xml:"scheduleType"`
}

// ScheduleTypes lists shift templates. includeDeleted is not implemented here,
// so deleted types come back regardless and callers must filter locally.
func (s *Service) ScheduleTypes(ctx context.Context) ([]ScheduleTypeDto, error) {
	var out scheduleTypeList
	if err := s.rest.GetXML(ctx, EndpointScheduleTypes, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

type attendanceTypeList struct {
	Items []AttendanceTypeDto `xml:"attendanceType"`
}

// AttendanceTypes lists the attendance categories.
func (s *Service) AttendanceTypes(ctx context.Context) ([]AttendanceTypeDto, error) {
	var out attendanceTypeList
	if err := s.rest.GetXML(ctx, EndpointAttendanceTypes, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

type availabilityList struct {
	Items []AvailabilityDto `xml:"availability"`
}

// Availability lists when employees may be scheduled.
//
// Unlike Schedules and Attendances, `to` is EXCLUSIVE here. Intervals are
// materialised for the whole range, so keep the window small.
func (s *Service) Availability(ctx context.Context, from, to time.Time, departmentIDs, roleIDs, userIDs []string) ([]AvailabilityDto, error) {
	q := url.Values{
		"from": {from.Format(rest.QueryV2)},
		"to":   {to.Format(rest.QueryV2)},
	}
	for _, d := range departmentIDs {
		q.Add("department", d)
	}
	for _, r := range roleIDs {
		q.Add("role", r)
	}
	for _, u := range userIDs {
		q.Add("user", u)
	}
	var out availabilityList
	if err := s.rest.GetXML(ctx, EndpointAvailability, q, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

type employeeList struct {
	Items []EmployeeDto `xml:"employee"`
}

// Employees lists active employees.
func (s *Service) Employees(ctx context.Context) ([]EmployeeDto, error) {
	return s.employeesAt(ctx, EndpointEmployees, nil)
}

// EmployeesByDepartment scopes the list to one department. On an RMS this
// filters nothing: department membership is only meaningful on a Chain.
func (s *Service) EmployeesByDepartment(ctx context.Context, departmentCode string) ([]EmployeeDto, error) {
	if strings.TrimSpace(departmentCode) == "" {
		return nil, fmt.Errorf("departmentCode is required")
	}
	return s.employeesAt(ctx, trimTemplate(EndpointEmployeesByDepartment)+url.PathEscape(departmentCode), nil)
}

// SearchEmployees matches on the documented regex fields; an unknown field name
// is ignored by iiko rather than rejected, so a typo silently widens the search.
func (s *Service) SearchEmployees(ctx context.Context, fields map[string]string) ([]EmployeeDto, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("at least one search field is required: an unfiltered employee list is the whole staff")
	}
	q := url.Values{}
	for k, v := range fields {
		q.Set(k, v)
	}
	return s.employeesAt(ctx, EndpointEmployeeSearch, q)
}

func (s *Service) employeesAt(ctx context.Context, path string, q url.Values) ([]EmployeeDto, error) {
	var out employeeList
	if err := s.rest.GetXML(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// EmployeeByID fetches one employee.
func (s *Service) EmployeeByID(ctx context.Context, id string) (*EmployeeDto, error) {
	return s.employee(ctx, trimTemplate(EndpointEmployeeByID), id, "id")
}

// EmployeeByCode fetches one employee by their code.
func (s *Service) EmployeeByCode(ctx context.Context, code string) (*EmployeeDto, error) {
	return s.employee(ctx, trimTemplate(EndpointEmployeeByCode), code, "code")
}

func (s *Service) employee(ctx context.Context, prefix, key, what string) (*EmployeeDto, error) {
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("employee %s is required", what)
	}
	var out EmployeeDto
	if err := s.rest.GetXML(ctx, prefix+url.PathEscape(key), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReplaceEmployee fully replaces an employee (PUT).
//
// This is a REPLACE, not a patch: every optional field left empty is CLEARED on
// the server. Read the employee first and send it back modified, or use
// UpdateEmployee for a partial change.
func (s *Service) ReplaceEmployee(ctx context.Context, id string, e EmployeeDto) (*EmployeeDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("employee id is required")
	}
	var out EmployeeDto
	if err := s.rest.PutXML(ctx, trimTemplate(EndpointEmployeeByID)+url.PathEscape(id), nil, e, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEmployee changes only the fields given (POST, form-urlencoded), leaving
// the rest as they are.
func (s *Service) UpdateEmployee(ctx context.Context, id string, fields map[string]string) (*EmployeeDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("employee id is required")
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}
	form := url.Values{}
	for k, v := range fields {
		form.Set(k, v)
	}
	data, err := s.rest.Do(ctx, http.MethodPost, trimTemplate(EndpointEmployeeByID)+url.PathEscape(id),
		nil, strings.NewReader(form.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return nil, err
	}
	var out EmployeeDto
	if err := xml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode updated employee: %w", err)
	}
	return &out, nil
}

// DeleteEmployee marks an employee deleted.
func (s *Service) DeleteEmployee(ctx context.Context, id string) error {
	return s.deleteByID(ctx, trimTemplate(EndpointEmployeeByID), id, "employee")
}

type roleList struct {
	Items []EmployeeRoleDto `xml:"role"`
}

// Roles lists job roles and their payment schemes.
func (s *Service) Roles(ctx context.Context, revisionFrom int) ([]EmployeeRoleDto, error) {
	q := url.Values{"revisionFrom": {fmt.Sprint(revisionFrom)}}
	var out roleList
	if err := s.rest.GetXML(ctx, EndpointEmployeeRoles, q, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

type salaryList struct {
	Items []SalaryDto `xml:"salary"`
}

// Salaries lists standing salaries for every employee.
func (s *Service) Salaries(ctx context.Context) ([]SalaryDto, error) {
	var out salaryList
	if err := s.rest.GetXML(ctx, EndpointSalary, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// SalaryByID reads one employee's salary, optionally as of a date. The date is
// a path segment, not a query parameter.
func (s *Service) SalaryByID(ctx context.Context, employeeID string, on *time.Time) ([]SalaryDto, error) {
	if strings.TrimSpace(employeeID) == "" {
		return nil, fmt.Errorf("employee id is required")
	}
	path := trimTemplate(EndpointSalaryByID) + url.PathEscape(employeeID)
	if on != nil {
		path += "/" + on.Format(rest.QueryV2)
	}
	var out salaryList
	if err := s.rest.GetXML(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// PaySalary records a salary payment on a date. It moves money, so it is a
// write in the plainest sense.
func (s *Service) PaySalary(ctx context.Context, employeeID string, on time.Time, payment string) error {
	if strings.TrimSpace(employeeID) == "" {
		return fmt.Errorf("employee id is required")
	}
	if strings.TrimSpace(payment) == "" {
		return fmt.Errorf("payment amount is required")
	}
	path := trimTemplate(EndpointSalaryOnDate) + url.PathEscape(employeeID) + "/" + on.Format(rest.QueryV2)
	_, err := s.rest.Do(ctx, http.MethodPost, path, url.Values{"payment": {payment}}, nil, "")
	return err
}

type waiterTeamList struct {
	Items []WaiterTeamDto `xml:"waiterTeam"`
}

// WaiterTeams lists serving teams.
//
// Teams are RMS-local: a Chain can read them but every write must go to the RMS
// that owns them.
func (s *Service) WaiterTeams(ctx context.Context) ([]WaiterTeamDto, error) {
	return s.teamsAt(ctx, EndpointWaiterTeams, nil)
}

// WaiterTeamsByDepartment scopes the list to one department.
func (s *Service) WaiterTeamsByDepartment(ctx context.Context, departmentID string) ([]WaiterTeamDto, error) {
	if strings.TrimSpace(departmentID) == "" {
		return nil, fmt.Errorf("departmentId is required")
	}
	return s.teamsAt(ctx, trimTemplate(EndpointWaiterTeamsByDepartment)+url.PathEscape(departmentID), nil)
}

// SearchWaiterTeams matches teams on the documented regex fields.
func (s *Service) SearchWaiterTeams(ctx context.Context, fields map[string]string, includeDeleted bool) ([]WaiterTeamDto, error) {
	q := url.Values{"includeDeleted": {rest.BoolStr(includeDeleted)}}
	for k, v := range fields {
		q.Set(k, v)
	}
	return s.teamsAt(ctx, EndpointWaiterTeamSearch, q)
}

func (s *Service) teamsAt(ctx context.Context, path string, q url.Values) ([]WaiterTeamDto, error) {
	var out waiterTeamList
	if err := s.rest.GetXML(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// WaiterTeamByID fetches one team.
func (s *Service) WaiterTeamByID(ctx context.Context, id string) (*WaiterTeamDto, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("team id is required")
	}
	var out WaiterTeamDto
	if err := s.rest.GetXML(ctx, trimTemplate(EndpointWaiterTeamByID)+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReplaceWaiterTeamByCode creates or replaces a team by code. A code already in
// use answers 409.
func (s *Service) ReplaceWaiterTeamByCode(ctx context.Context, code string, team WaiterTeamDto) (*WaiterTeamDto, error) {
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("team code is required")
	}
	var out WaiterTeamDto
	if err := s.rest.PutXML(ctx, trimTemplate(EndpointWaiterTeamByCode)+url.PathEscape(code), nil, team, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteWaiterTeam removes a team.
//
// The assignment map is NOT garbage-collected when a team goes: leftover entries
// keep pointing at it, and removing an employee from a team widens rather than
// narrows what they can see. Re-read Assignments after any delete.
func (s *Service) DeleteWaiterTeam(ctx context.Context, id string) error {
	return s.deleteByID(ctx, trimTemplate(EndpointWaiterTeamByID), id, "team")
}

// Assignments reads the employee-to-team map across all departments.
func (s *Service) Assignments(ctx context.Context) (*WaiterTeamAssignmentsDto, error) {
	var out WaiterTeamAssignmentsDto
	if err := s.rest.GetXML(ctx, EndpointWaiterTeamAssignments, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AssignmentsByDepartment reads the map for one department.
func (s *Service) AssignmentsByDepartment(ctx context.Context, departmentID string) (*WaiterTeamAssignmentsDto, error) {
	if strings.TrimSpace(departmentID) == "" {
		return nil, fmt.Errorf("departmentId is required")
	}
	var out WaiterTeamAssignmentsDto
	if err := s.rest.GetXML(ctx, trimTemplate(EndpointWaiterTeamAssignmentsByDeptID)+url.PathEscape(departmentID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReplaceAssignments replaces the whole employee-to-team map. It is a replace:
// an employee left out is unassigned, which widens their order visibility.
func (s *Service) ReplaceAssignments(ctx context.Context, a WaiterTeamAssignmentsDto) error {
	return s.rest.PutXML(ctx, EndpointWaiterTeamAssignments, nil, a, nil)
}
