// Package employees covers staff, positions, personal shifts, courier locations and attendance.
package employees

import (
	"context"

	"github.com/Nap20192/iiko-go/iikocloud/gen"
	"github.com/Nap20192/iiko-go/iikocloud/rest"
)

// Service is the employees half of iikoCloud.
type Service struct{ c *rest.Client }

// New wires a service onto a transport; the facade is the usual caller.
func New(c *rest.Client) *Service { return &Service{c: c} }

// ClockinShift open personal session.
func (s *Service) ClockinShift(ctx context.Context, req gen.OpenPersonalSessionRequest) (*gen.ChangePersonalSessionResponse, error) {
	return rest.Call[gen.ChangePersonalSessionResponse](ctx, s.c, clockinShift, req)
}

// ClockoutShift close personal session.
func (s *Service) ClockoutShift(ctx context.Context, req gen.ClosePersonalSessionRequest) (*gen.ChangePersonalSessionResponse, error) {
	return rest.Call[gen.ChangePersonalSessionResponse](ctx, s.c, clockoutShift, req)
}

// CreateAttendance create employee attendance.
func (s *Service) CreateAttendance(ctx context.Context, req gen.AttendanceCreateRequest) error {
	return s.c.Post(ctx, createAttendance, req, nil)
}

// CreateEmployee create employee.
func (s *Service) CreateEmployee(ctx context.Context, req gen.EmployeeCreateRequest) error {
	return s.c.Post(ctx, createEmployee, req, nil)
}

// DeleteAttendance delete employee attendance.
func (s *Service) DeleteAttendance(ctx context.Context, req gen.AttendanceDeleteRequest) (*gen.AttendanceDeleteResponse, error) {
	return rest.Call[gen.AttendanceDeleteResponse](ctx, s.c, deleteAttendance, req)
}

// FireEmployee fire employees.
func (s *Service) FireEmployee(ctx context.Context, req gen.EmployeeFireRequest) (*gen.EmployeeFireResponse, error) {
	return rest.Call[gen.EmployeeFireResponse](ctx, s.c, fireEmployee, req)
}

// GetCouriers returns list of all employees which are delivery drivers in specified restaurants.
func (s *Service) GetCouriers(ctx context.Context, req gen.CouriersRequest) (*gen.EmployeesResponse, error) {
	return rest.Call[gen.EmployeesResponse](ctx, s.c, getCouriers, req)
}

// GetCouriersActiveLocation returns list of all active (courier session is opened) courier's locations which are delivery drivers in specified restaurants.
func (s *Service) GetCouriersActiveLocation(ctx context.Context, req gen.CouriersRequest) (*gen.ActiveCourierLocationsResponse, error) {
	return rest.Call[gen.ActiveCourierLocationsResponse](ctx, s.c, getCouriersActiveLocation, req)
}

// GetCouriersActiveLocationByTerminal returns list of all active (courier session is opened) courier's locations which are delivery drivers in specified restaurant and are clocked in on specified de….
func (s *Service) GetCouriersActiveLocationByTerminal(ctx context.Context, req gen.ActiveCourierLocationsByTerminalGroupRequest) (*gen.ActiveCourierLocationsResponse, error) {
	return rest.Call[gen.ActiveCourierLocationsResponse](ctx, s.c, getCouriersActiveLocationByTerminal, req)
}

// GetCouriersByRole returns list of all employees which are delivery drivers in specified restaurants, and checks whether each employee has passed role.
func (s *Service) GetCouriersByRole(ctx context.Context, req gen.CouriersAndCheckRoleRequest) (*gen.EmployeesWithRoleSignResponse, error) {
	return rest.Call[gen.EmployeesWithRoleSignResponse](ctx, s.c, getCouriersByRole, req)
}

// GetCouriersLocationsByTimeOffset method of obtaining drivers' coordinates history.
func (s *Service) GetCouriersLocationsByTimeOffset(ctx context.Context, req gen.CourierLocationsByTimeOffsetRequest) (*gen.CourierLocationsByTimeOffsetResponse, error) {
	return rest.Call[gen.CourierLocationsByTimeOffsetResponse](ctx, s.c, getCouriersLocationsByTimeOffset, req)
}

// GetEmployee get employee.
func (s *Service) GetEmployee(ctx context.Context, req gen.EmployeeGetRequest) (*gen.EmployeeResponse, error) {
	return rest.Call[gen.EmployeeResponse](ctx, s.c, getEmployee, req)
}

// GetEmployeeInfo returns employee info.
func (s *Service) GetEmployeeInfo(ctx context.Context, req gen.EmployeeInfoRequest) (*gen.EmployeeInfoResponse, error) {
	return rest.Call[gen.EmployeeInfoResponse](ctx, s.c, getEmployeeInfo, req)
}

// GetPositions get employee position by identifier.
func (s *Service) GetPositions(ctx context.Context, req gen.GetPositionRequest) (*gen.EmployeePositionResponse, error) {
	return rest.Call[gen.EmployeePositionResponse](ctx, s.c, getPositions, req)
}

// GetShiftIsOpen check if personal session is open.
func (s *Service) GetShiftIsOpen(ctx context.Context, req gen.GetPersonalSessionInfoRequest) (*gen.GetPersonalSessionInfoResponse, error) {
	return rest.Call[gen.GetPersonalSessionInfoResponse](ctx, s.c, getShiftIsOpen, req)
}

// GetShiftsByCourier get terminal groups where employee session is opened.
func (s *Service) GetShiftsByCourier(ctx context.Context, req gen.GetTerminalGroupsOfEmployeeRequest) (*gen.GetTerminalGroupsOfEmployeeResponse, error) {
	return rest.Call[gen.GetTerminalGroupsOfEmployeeResponse](ctx, s.c, getShiftsByCourier, req)
}

// ListAttendance list employee attendances.
func (s *Service) ListAttendance(ctx context.Context, req gen.AttendanceListRequest) (*gen.AttendanceListResponse, error) {
	return rest.Call[gen.AttendanceListResponse](ctx, s.c, listAttendance, req)
}

// ListAttendanceType list attendance types.
func (s *Service) ListAttendanceType(ctx context.Context, req gen.AttendanceTypeListRequest) (*gen.AttendanceTypeListResponse, error) {
	return rest.Call[gen.AttendanceTypeListResponse](ctx, s.c, listAttendanceType, req)
}

// ListEmployee list of employees.
func (s *Service) ListEmployee(ctx context.Context, req gen.EmployeeListRequest) (*gen.SwaggerEmployeeListResponse, error) {
	return rest.Call[gen.SwaggerEmployeeListResponse](ctx, s.c, listEmployee, req)
}

// ListPositions list employee positions.
func (s *Service) ListPositions(ctx context.Context, req gen.ListPositionsRequest) (*gen.ListPositionsResponse, error) {
	return rest.Call[gen.ListPositionsResponse](ctx, s.c, listPositions, req)
}

// RestoreEmployee restore employee.
func (s *Service) RestoreEmployee(ctx context.Context, req gen.EmployeeRestoreRequest) (*gen.EmployeeRestoreResponse, error) {
	return rest.Call[gen.EmployeeRestoreResponse](ctx, s.c, restoreEmployee, req)
}

// UpdateAttendance update employee attendance.
func (s *Service) UpdateAttendance(ctx context.Context, req gen.AttendanceUpdateRequest) (*gen.AttendanceUpdateResponse, error) {
	return rest.Call[gen.AttendanceUpdateResponse](ctx, s.c, updateAttendance, req)
}

// UpdateEmployee update employee.
func (s *Service) UpdateEmployee(ctx context.Context, req gen.EmployeeUpdateRequest) (*gen.EmployeeResponse, error) {
	return rest.Call[gen.EmployeeResponse](ctx, s.c, updateEmployee, req)
}
