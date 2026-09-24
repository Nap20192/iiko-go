// Package employees paths. Every path is a constant here; none appears in a method body.

package employees

const (
	clockinShift                        = "/api/1/employees/shift/clockin"
	clockoutShift                       = "/api/1/employees/shift/clockout"
	createAttendance                    = "/api/employees/v1/attendance/create"
	createEmployee                      = "/api/employees/v1/employee/create"
	deleteAttendance                    = "/api/employees/v1/attendance/delete"
	fireEmployee                        = "/api/employees/v1/employee/fire"
	getCouriers                         = "/api/1/employees/couriers"
	getCouriersActiveLocation           = "/api/1/employees/couriers/active_location"
	getCouriersActiveLocationByTerminal = "/api/1/employees/couriers/active_location/by_terminal"
	getCouriersByRole                   = "/api/1/employees/couriers/by_role"
	getCouriersLocationsByTimeOffset    = "/api/1/employees/couriers/locations/by_time_offset"
	getEmployee                         = "/api/employees/v1/employee/get"
	getEmployeeInfo                     = "/api/1/employees/info"
	getPositions                        = "/api/employees/v1/positions/get"
	getShiftIsOpen                      = "/api/1/employees/shift/is_open"
	getShiftsByCourier                  = "/api/1/employees/shifts/by_courier"
	listAttendance                      = "/api/employees/v1/attendance/list"
	listAttendanceType                  = "/api/employees/v1/attendance-type/list"
	listEmployee                        = "/api/employees/v1/employee/list"
	listPositions                       = "/api/employees/v1/positions/list"
	restoreEmployee                     = "/api/employees/v1/employee/restore"
	updateAttendance                    = "/api/employees/v1/attendance/update"
	updateEmployee                      = "/api/employees/v1/employee/update"
)
