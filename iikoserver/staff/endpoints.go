package staff

// Employee, schedule, attendance and waiter-team paths, one per documented
// endpoint. Schedule and attendance each expose six read shapes that differ
// only in scope; the service picks one from the filter it is given.
const (
	// Employees — XML, iiko 4.0.
	EndpointEmployees             = "/api/employees"
	EndpointEmployeeByID          = "/api/employees/byId/{uuid}"   // GET read, PUT full replace, POST partial update, DELETE
	EndpointEmployeeByCode        = "/api/employees/byCode/{code}" // GET read, PUT full replace
	EndpointEmployeesByDepartment = "/api/employees/byDepartment/{departmentCode}"
	EndpointEmployeeSearch        = "/api/employees/search"
	EndpointEmployeeRoles         = "/api/employees/roles"

	// Employee writes. PUT fully replaces and clears omitted optional fields;
	// POST updates only the fields sent, form-urlencoded.

	// Salary.
	EndpointSalary       = "/api/employees/salary"
	EndpointSalaryByID   = "/api/employees/salary/byId/{uuid}"
	EndpointSalaryOnDate = "/api/employees/salary/byId/{uuid}/{date}" // GET read, POST records a payment

	// Schedules. `to` is INCLUSIVE here, which the docs flag as an exception.
	EndpointSchedule                         = "/api/employees/schedule/"
	EndpointScheduleByEmployee               = "/api/employees/schedule/byEmployee/{uuid}/"
	EndpointScheduleByDepartmentCode         = "/api/employees/schedule/byDepartment/{code}/"
	EndpointScheduleByDepartmentCodeEmployee = "/api/employees/schedule/byDepartment/{code}/byEmployee/{uuid}/"
	EndpointScheduleByDepartmentID           = "/api/employees/schedule/department/{id}/"
	EndpointScheduleByDepartmentIDEmployee   = "/api/employees/schedule/department/{id}/byEmployee/{uuid}/"
	EndpointScheduleTypes                    = "/api/employees/schedule/types" // includeDeleted is NOT implemented
	EndpointScheduleCreate                   = "/api/employees/schedule/create"
	EndpointScheduleUpdate                   = "/api/employees/schedule/update" // MAY CHANGE the shift's id
	EndpointScheduleDelete                   = "/api/employees/schedule/byId/{uuid}"

	// Attendances, mirroring the schedule shapes. `to` is INCLUSIVE here too.
	EndpointAttendance                         = "/api/employees/attendance"
	EndpointAttendanceByEmployee               = "/api/employees/attendance/byEmployee/{uuid}/"
	EndpointAttendanceByDepartmentCode         = "/api/employees/attendance/byDepartment/{code}/"
	EndpointAttendanceByDepartmentCodeEmployee = "/api/employees/attendance/byDepartment/{code}/byEmployee/{uuid}/"
	EndpointAttendanceByDepartmentID           = "/api/employees/attendance/department/{id}/"
	EndpointAttendanceByDepartmentIDEmployee   = "/api/employees/attendance/department/{id}/byEmployee/{uuid}/"
	EndpointAttendanceTypes                    = "/api/employees/attendance/types"
	EndpointAttendanceCreate                   = "/api/employees/attendance/create"
	EndpointAttendanceUpdate                   = "/api/employees/attendance/update" // MAY CHANGE the entry's id
	EndpointAttendanceDelete                   = "/api/employees/attendance/byId/{uuid}"

	// Availability. Unlike schedule and attendance, `to` is EXCLUSIVE.
	EndpointAvailability = "/api/employees/availability/list"

	// Waiter teams, iiko 8.1. Writes land on an RMS only.
	EndpointWaiterTeams                   = "/api/employees/waiterTeams"
	EndpointWaiterTeamByID                = "/api/employees/waiterTeams/byId/{uuid}"   // GET read, PUT replace, POST update, DELETE
	EndpointWaiterTeamByCode              = "/api/employees/waiterTeams/byCode/{code}" // PUT replace; GET is the verb-inferred read
	EndpointWaiterTeamsByDepartment       = "/api/employees/waiterTeams/byDepartment/{uuid}"
	EndpointWaiterTeamSearch              = "/api/employees/waiterTeams/search"
	EndpointWaiterTeamAssignments         = "/api/employees/waiterTeams/assignments" // GET read, PUT replace the whole map
	EndpointWaiterTeamAssignmentsByDeptID = "/api/employees/waiterTeams/assignments/byDepartment/{id}"
)
