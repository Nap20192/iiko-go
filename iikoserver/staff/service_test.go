package staff

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Nap20192/iiko-go/iikoserver/rest"
	"github.com/Nap20192/iiko-go/iikoserver/rest/resttest"
)

func newFake(t *testing.T, h http.HandlerFunc) (*Service, *rest.Client) {
	t.Helper()
	c := resttest.NewFake(t, h)
	return New(c), c
}

func day(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := rest.ParseDay(s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return d
}

// TRAP 1. The docs flag this as an exception to the rest of the API: on
// schedules and attendances `to` is INCLUSIVE, so a caller asking for a single
// day must get that day, not an empty range.
func TestShiftWindowIsInclusiveAtBothEnds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		read         func(*Service, ShiftFilter) error
		wantPathPart string
	}{
		{
			name:         "schedules",
			read:         func(s *Service, f ShiftFilter) error { _, err := s.Schedules(context.Background(), f); return err },
			wantPathPart: "/schedule/",
		},
		{
			name:         "attendances",
			read:         func(s *Service, f ShiftFilter) error { _, err := s.Attendances(context.Background(), f); return err },
			wantPathPart: "/attendance",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotQ url.Values
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotQ, gotPath = r.URL.Query(), r.URL.Path
				_, _ = w.Write([]byte(`<r></r>`))
			})
			one := day(t, "2026-03-05")
			if err := tt.read(s, ShiftFilter{From: one, To: one}); err != nil {
				t.Fatalf("read: %v", err)
			}
			if !strings.Contains(gotPath, tt.wantPathPart) {
				t.Errorf("path %q must contain %q", gotPath, tt.wantPathPart)
			}
			// A single day must be sent as itself at both ends, not widened.
			if gotQ.Get("from") != "2026-03-05" || gotQ.Get("to") != "2026-03-05" {
				t.Errorf("one inclusive day must send from=to=2026-03-05, got from=%q to=%q",
					gotQ.Get("from"), gotQ.Get("to"))
			}
		})
	}
}

// TRAP 2. withPaymentDetails=true fails outright when the employee still has an
// unclosed earlier attendance. The failure has to reach the caller with the
// reason, not be swallowed into an empty list.
func TestWithPaymentDetailsSurfacesUnclosedAttendanceError(t *testing.T) {
	t.Parallel()
	var gotQ url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Cannot compute payment: employee has unclosed attendance"))
	})
	one := day(t, "2026-03-05")
	_, err := s.Schedules(context.Background(), ShiftFilter{From: one, To: one, WithPaymentDetails: true})
	if err == nil {
		t.Fatal("an unclosed earlier attendance must surface, not yield an empty schedule")
	}
	if !strings.Contains(err.Error(), "unclosed attendance") {
		t.Errorf("the server's own reason must reach the caller, got %v", err)
	}
	if gotQ.Get("withPaymentDetails") != "true" {
		t.Errorf("withPaymentDetails must be sent explicitly, got %q", gotQ.Get("withPaymentDetails"))
	}
}

// withPaymentDetails is off by default, and sent explicitly either way.
func TestShiftReadsSendFlagsExplicitly(t *testing.T) {
	t.Parallel()
	var gotQ url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query()
		_, _ = w.Write([]byte(`<r></r>`))
	})
	one := day(t, "2026-03-05")
	if _, err := s.Schedules(context.Background(), ShiftFilter{From: one, To: one}); err != nil {
		t.Fatalf("read: %v", err)
	}
	if gotQ.Get("withPaymentDetails") != "false" {
		t.Errorf("withPaymentDetails must be explicit even when false, got %q", gotQ.Get("withPaymentDetails"))
	}
	if gotQ.Get("revisionFrom") != "-1" {
		t.Errorf("revisionFrom defaults to -1 and is sent explicitly, got %q", gotQ.Get("revisionFrom"))
	}
}

// TRAP 3. The docs warn that updating a shift may change its id, so the caller
// must be handed the id the server came back with rather than reusing its own.
func TestScheduleUpdateReturnsTheServersIDNotTheSubmittedOne(t *testing.T) {
	t.Parallel()
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		// iiko answers with the stored shift, whose id it may have replaced.
		_, _ = w.Write([]byte(`<schedules><schedule><id>new-id</id><employeeId>e-1</employeeId></schedule></schedules>`))
	})
	got, err := s.UpdateSchedule(context.Background(), ScheduleDto{ID: "old-id", EmployeeID: "e-1"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !strings.HasSuffix(gotPath, EndpointScheduleUpdate) {
		t.Errorf("wrong path %q", gotPath)
	}
	if got.ID != "new-id" {
		t.Errorf("update returned id %q; iiko may replace the id, so the caller must get the stored one, never the submitted one", got.ID)
	}
}

func TestShiftFilterPicksTheScopedPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		filter ShiftFilter
		want   string
	}{
		{name: "unscoped", filter: ShiftFilter{}, want: "/api/employees/schedule/"},
		{name: "by employee", filter: ShiftFilter{EmployeeID: "e-1"}, want: "/api/employees/schedule/byEmployee/e-1/"},
		{name: "by department code", filter: ShiftFilter{DepartmentCode: "D1"}, want: "/api/employees/schedule/byDepartment/D1/"},
		{name: "by department code and employee", filter: ShiftFilter{DepartmentCode: "D1", EmployeeID: "e-1"},
			want: "/api/employees/schedule/byDepartment/D1/byEmployee/e-1/"},
		{name: "by department id", filter: ShiftFilter{DepartmentID: "d-1"}, want: "/api/employees/schedule/department/d-1/"},
		{name: "by department id and employee", filter: ShiftFilter{DepartmentID: "d-1", EmployeeID: "e-1"},
			want: "/api/employees/schedule/department/d-1/byEmployee/e-1/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = w.Write([]byte(`<r></r>`))
			})
			f := tt.filter
			f.From, f.To = day(t, "2026-03-01"), day(t, "2026-03-02")
			if _, err := s.Schedules(context.Background(), f); err != nil {
				t.Fatalf("read: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.want) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.want)
			}
		})
	}
}

// Scoping by both a code and an id is ambiguous; iiko has no such endpoint.
func TestShiftFilterRejectsTwoDepartmentScopes(t *testing.T) {
	t.Parallel()
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not reach iiko with an ambiguous scope")
		w.WriteHeader(http.StatusInternalServerError)
	})
	one := day(t, "2026-03-05")
	_, err := s.Schedules(context.Background(), ShiftFilter{
		From: one, To: one, DepartmentCode: "D1", DepartmentID: "d-1",
	})
	if err == nil || !strings.Contains(err.Error(), "department_code or department_id") {
		t.Errorf("want an ambiguous-scope error, got %v", err)
	}
}

// PUT replaces and POST patches. Confusing them silently wipes an employee's
// optional fields, so the verbs and bodies are asserted, not assumed.
func TestEmployeeReplaceAndUpdateUseDifferentVerbsAndBodies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		call       func(*Service) error
		wantMethod string
		wantCT     string
		wantInBody string
	}{
		{
			name: "replace is a PUT carrying the whole entity as XML",
			call: func(s *Service) error {
				_, err := s.ReplaceEmployee(context.Background(), "e-1", EmployeeDto{Name: "Ann"})
				return err
			},
			wantMethod: http.MethodPut, wantCT: "xml", wantInBody: "<name>Ann</name>",
		},
		{
			name: "update is a form-urlencoded POST of only the named fields",
			call: func(s *Service) error {
				_, err := s.UpdateEmployee(context.Background(), "e-1", map[string]string{"name": "Ann"})
				return err
			},
			wantMethod: http.MethodPost, wantCT: "form-urlencoded", wantInBody: "name=Ann",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotMethod, gotCT, gotBody, gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				b := make([]byte, r.ContentLength)
				_, _ = r.Body.Read(b)
				gotMethod, gotCT, gotBody, gotPath = r.Method, r.Header.Get("Content-Type"), string(b), r.URL.Path
				_, _ = w.Write([]byte(`<employee><id>e-1</id><name>Ann</name></employee>`))
			})
			if err := tt.call(s); err != nil {
				t.Fatalf("call: %v", err)
			}
			if gotMethod != tt.wantMethod {
				t.Errorf("method = %q, want %q", gotMethod, tt.wantMethod)
			}
			if !strings.Contains(gotCT, tt.wantCT) {
				t.Errorf("Content-Type = %q, want %q", gotCT, tt.wantCT)
			}
			if !strings.Contains(gotBody, tt.wantInBody) {
				t.Errorf("body = %q, want %q in it", gotBody, tt.wantInBody)
			}
			if !strings.HasSuffix(gotPath, "/api/employees/byId/e-1") {
				t.Errorf("path = %q", gotPath)
			}
		})
	}
}

// Availability is the one shift-shaped read whose `to` is EXCLUSIVE, and its
// three filters repeat rather than joining.
func TestAvailabilityRepeatsFiltersAndKeepsItsOwnWindow(t *testing.T) {
	t.Parallel()
	var gotQ url.Values
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotQ, gotPath = r.URL.Query(), r.URL.Path
		_, _ = w.Write([]byte(`<availabilities></availabilities>`))
	})
	if _, err := s.Availability(context.Background(), day(t, "2026-03-01"), day(t, "2026-03-08"),
		[]string{"d1", "d2"}, []string{"r1"}, nil); err != nil {
		t.Fatalf("availability: %v", err)
	}
	if !strings.HasSuffix(gotPath, EndpointAvailability) {
		t.Errorf("path = %q", gotPath)
	}
	if got := gotQ["department"]; len(got) != 2 || got[0] != "d1" || got[1] != "d2" {
		t.Errorf("department must repeat per value, got %v", got)
	}
	if got := gotQ["role"]; len(got) != 1 {
		t.Errorf("role must repeat per value, got %v", got)
	}
	if _, present := gotQ["user"]; present {
		t.Error("an empty filter must be omitted, not sent blank")
	}
	if gotQ.Get("to") != "2026-03-08" {
		t.Errorf("to = %q; availability's `to` is exclusive and is sent as given", gotQ.Get("to"))
	}
}

// The salary date is a path segment; sending it as a query silently reads today.
func TestSalaryDateIsAPathSegment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		on   bool
		want string
	}{
		{name: "without a date", want: "/api/employees/salary/byId/e-1"},
		{name: "as of a date", on: true, want: "/api/employees/salary/byId/e-1/2026-03-05"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			var gotQ url.Values
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotQ = r.URL.Path, r.URL.Query()
				_, _ = w.Write([]byte(`<salaries></salaries>`))
			})
			var on *time.Time
			if tt.on {
				d := day(t, "2026-03-05")
				on = &d
			}
			if _, err := s.SalaryByID(context.Background(), "e-1", on); err != nil {
				t.Fatalf("salary: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.want) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.want)
			}
			if gotQ.Get("date") != "" {
				t.Error("the date belongs in the path, not the query")
			}
		})
	}
}

func TestStaffReadsRejectMissingIdentifiers(t *testing.T) {
	t.Parallel()
	s, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("must not reach iiko without an identifier")
		w.WriteHeader(http.StatusInternalServerError)
	})
	ctx := context.Background()
	checks := map[string]error{
		"EmployeeByID":    func() error { _, err := s.EmployeeByID(ctx, " "); return err }(),
		"EmployeeByCode":  func() error { _, err := s.EmployeeByCode(ctx, ""); return err }(),
		"SalaryByID":      func() error { _, err := s.SalaryByID(ctx, "", nil); return err }(),
		"DeleteEmployee":  s.DeleteEmployee(ctx, ""),
		"DeleteSchedule":  s.DeleteSchedule(ctx, ""),
		"SearchEmployees": func() error { _, err := s.SearchEmployees(ctx, nil); return err }(),
		"PaySalary":       s.PaySalary(ctx, "e-1", day(t, "2026-03-05"), ""),
	}
	for name, err := range checks {
		if err == nil {
			t.Errorf("%s must reject a missing identifier before spending a request", name)
		}
	}
}

// Every remaining reader, asserted where they can actually go wrong: the path
// they hit and that the XML element name they expect is the one iiko sends.
func TestStaffReadersHitTheRightPathAndDecode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		body     string
		call     func(*Service) (int, error)
		wantPath string
	}{
		{
			name: "employees", wantPath: "/api/employees",
			body: `<employees><employee><id>e-1</id></employee></employees>`,
			call: func(s *Service) (int, error) { v, err := s.Employees(context.Background()); return len(v), err },
		},
		{
			name: "employees by department", wantPath: "/api/employees/byDepartment/D1",
			body: `<employees><employee><id>e-1</id></employee></employees>`,
			call: func(s *Service) (int, error) {
				v, err := s.EmployeesByDepartment(context.Background(), "D1")
				return len(v), err
			},
		},
		{
			name: "roles", wantPath: "/api/employees/roles",
			body: `<roles><role><id>r-1</id><name>Waiter</name></role></roles>`,
			call: func(s *Service) (int, error) { v, err := s.Roles(context.Background(), -1); return len(v), err },
		},
		{
			name: "salaries", wantPath: "/api/employees/salary",
			body: `<salaries><salary><employeeId>e-1</employeeId></salary></salaries>`,
			call: func(s *Service) (int, error) { v, err := s.Salaries(context.Background()); return len(v), err },
		},
		{
			name: "schedule types", wantPath: "/api/employees/schedule/types",
			body: `<scheduleTypes><scheduleType><id>st-1</id></scheduleType></scheduleTypes>`,
			call: func(s *Service) (int, error) { v, err := s.ScheduleTypes(context.Background()); return len(v), err },
		},
		{
			name: "attendance types", wantPath: "/api/employees/attendance/types",
			body: `<attendanceTypes><attendanceType><id>at-1</id></attendanceType></attendanceTypes>`,
			call: func(s *Service) (int, error) { v, err := s.AttendanceTypes(context.Background()); return len(v), err },
		},
		{
			name: "waiter teams", wantPath: "/api/employees/waiterTeams",
			body: `<waiterTeams><waiterTeam><id>t-1</id></waiterTeam></waiterTeams>`,
			call: func(s *Service) (int, error) { v, err := s.WaiterTeams(context.Background()); return len(v), err },
		},
		{
			name: "waiter teams by department", wantPath: "/api/employees/waiterTeams/byDepartment/d-1",
			body: `<waiterTeams><waiterTeam><id>t-1</id></waiterTeam></waiterTeams>`,
			call: func(s *Service) (int, error) {
				v, err := s.WaiterTeamsByDepartment(context.Background(), "d-1")
				return len(v), err
			},
		},
		{
			name: "waiter team search", wantPath: "/api/employees/waiterTeams/search",
			body: `<waiterTeams><waiterTeam><id>t-1</id></waiterTeam></waiterTeams>`,
			call: func(s *Service) (int, error) {
				v, err := s.SearchWaiterTeams(context.Background(), map[string]string{"name": "A"}, false)
				return len(v), err
			},
		},
		{
			name: "assignments", wantPath: "/api/employees/waiterTeams/assignments",
			body: `<assignments><assignment><employeeId>e-1</employeeId></assignment></assignments>`,
			call: func(s *Service) (int, error) {
				v, err := s.Assignments(context.Background())
				if err != nil {
					return 0, err
				}
				return len(v.Assignment), nil
			},
		},
		{
			name: "assignments by department", wantPath: "/api/employees/waiterTeams/assignments/byDepartment/d-1",
			body: `<assignments><assignment><employeeId>e-1</employeeId></assignment></assignments>`,
			call: func(s *Service) (int, error) {
				v, err := s.AssignmentsByDepartment(context.Background(), "d-1")
				if err != nil {
					return 0, err
				}
				return len(v.Assignment), nil
			},
		},
		{
			name: "employee search", wantPath: "/api/employees/search",
			body: `<employees><employee><id>e-1</id></employee></employees>`,
			call: func(s *Service) (int, error) {
				v, err := s.SearchEmployees(context.Background(), map[string]string{"lastName": "Ivanov"})
				return len(v), err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				_, _ = w.Write([]byte(tt.body))
			})
			n, err := tt.call(s)
			if err != nil {
				t.Fatalf("call: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.wantPath) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.wantPath)
			}
			if n != 1 {
				t.Errorf("decoded %d items, want 1 — the XML element name probably does not match what iiko sends", n)
			}
		})
	}
}

func TestStaffWritersHitTheRightPathAndVerb(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		call       func(*Service) error
		wantPath   string
		wantMethod string
		body       string // the writers that read a body back need their own shape
	}{
		{
			name: "delete schedule", wantPath: "/api/employees/schedule/byId/s-1", wantMethod: http.MethodDelete,
			call: func(s *Service) error { return s.DeleteSchedule(context.Background(), "s-1") },
		},
		{
			name: "delete attendance", wantPath: "/api/employees/attendance/byId/a-1", wantMethod: http.MethodDelete,
			call: func(s *Service) error { return s.DeleteAttendance(context.Background(), "a-1") },
		},
		{
			name: "delete employee", wantPath: "/api/employees/byId/e-1", wantMethod: http.MethodDelete,
			call: func(s *Service) error { return s.DeleteEmployee(context.Background(), "e-1") },
		},
		{
			name: "delete waiter team", wantPath: "/api/employees/waiterTeams/byId/t-1", wantMethod: http.MethodDelete,
			call: func(s *Service) error { return s.DeleteWaiterTeam(context.Background(), "t-1") },
		},
		{
			name: "pay salary", wantPath: "/api/employees/salary/byId/e-1/2026-03-05", wantMethod: http.MethodPost,
			call: func(s *Service) error {
				return s.PaySalary(context.Background(), "e-1", mustDay("2026-03-05"), "1000")
			},
		},
		{
			name: "replace assignments", wantPath: "/api/employees/waiterTeams/assignments", wantMethod: http.MethodPut,
			call: func(s *Service) error {
				return s.ReplaceAssignments(context.Background(), WaiterTeamAssignmentsDto{DepartmentID: "d-1"})
			},
		},
		{
			name: "replace team by code", wantPath: "/api/employees/waiterTeams/byCode/T1", wantMethod: http.MethodPut,
			call: func(s *Service) error {
				_, err := s.ReplaceWaiterTeamByCode(context.Background(), "T1", WaiterTeamDto{Name: "A"})
				return err
			},
		},
		{
			name: "create attendance", wantPath: "/api/employees/attendance/create", wantMethod: http.MethodPost,
			body: `<attendances><attendance><id>x</id><employeeId>e-1</employeeId></attendance></attendances>`,
			call: func(s *Service) error {
				_, err := s.CreateAttendance(context.Background(), AttendanceDto{EmployeeID: "e-1"})
				return err
			},
		},
		{
			name: "create schedule", wantPath: "/api/employees/schedule/create", wantMethod: http.MethodPost,
			call: func(s *Service) error {
				_, err := s.CreateSchedule(context.Background(), ScheduleDto{EmployeeID: "e-1"})
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath, gotMethod string
			body := tt.body
			if body == "" {
				body = `<schedules><schedule><id>x</id><employeeId>e-1</employeeId></schedule></schedules>`
			}
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotMethod = r.URL.Path, r.Method
				_, _ = w.Write([]byte(body))
			})
			if err := tt.call(s); err != nil {
				t.Fatalf("call: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.wantPath) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.wantPath)
			}
			if gotMethod != tt.wantMethod {
				t.Errorf("method = %q, want %q", gotMethod, tt.wantMethod)
			}
		})
	}
}

// mustDay is for table entries built before a *testing.T is in scope.
func mustDay(s string) time.Time {
	d, err := rest.ParseDay(s)
	if err != nil {
		panic(err)
	}
	return d
}

// The mutating-id warning covers attendances as well as schedules.
func TestAttendanceUpdateReturnsTheServersID(t *testing.T) {
	t.Parallel()
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`<attendances><attendance><id>new-id</id><employeeId>e-1</employeeId></attendance></attendances>`))
	})
	got, err := s.UpdateAttendance(context.Background(), AttendanceDto{ID: "old-id", EmployeeID: "e-1"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !strings.HasSuffix(gotPath, EndpointAttendanceUpdate) {
		t.Errorf("path = %q", gotPath)
	}
	if got.ID != "new-id" {
		t.Errorf("id = %q; iiko may replace it on update, so the stored id is the only safe one", got.ID)
	}
}

func TestWaiterTeamByIDFetchesOne(t *testing.T) {
	t.Parallel()
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`<waiterTeam><id>t-1</id><name>Hall</name></waiterTeam>`))
	})
	got, err := s.WaiterTeamByID(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("byId: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/api/employees/waiterTeams/byId/t-1") {
		t.Errorf("path = %q", gotPath)
	}
	if got.Name != "Hall" {
		t.Errorf("decoded %+v", got)
	}
}
