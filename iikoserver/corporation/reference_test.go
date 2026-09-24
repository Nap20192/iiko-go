package corporation

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Settings is documented at two paths on the same page. Like accounts/list, the
// /v2/ spelling is tried first and a 404 falls back — but a 403 must not, or a
// permission problem is retried as if it were a wrong path.
func TestSettingsFallsBackOn404ButNotOn403(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		firstCode int
		wantCalls int
		wantErr   bool
	}{
		{name: "v2 answers", firstCode: http.StatusOK, wantCalls: 1},
		{name: "404 falls back to the v1 spelling", firstCode: http.StatusNotFound, wantCalls: 2},
		{name: "403 does not fall back", firstCode: http.StatusForbidden, wantCalls: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var calls int
			var paths []string
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				paths = append(paths, r.URL.Path)
				if calls == 1 && tt.firstCode != http.StatusOK {
					w.WriteHeader(tt.firstCode)
					_, _ = w.Write([]byte("nope"))
					return
				}
				_, _ = w.Write([]byte(`{"vatAccounting":"VAT_INCLUDED_IN_PRICE"}`))
			})
			got, err := s.Settings(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("a 403 is a permission problem, not a wrong path")
				}
			} else if err != nil {
				t.Fatalf("settings: %v", err)
			} else if got.VatAccounting != "VAT_INCLUDED_IN_PRICE" {
				t.Errorf("decoded %+v", got)
			}
			if calls != tt.wantCalls {
				t.Errorf("made %d calls, want %d (paths %v)", calls, tt.wantCalls, paths)
			}
			if !strings.Contains(paths[0], "/v2/") {
				t.Errorf("the /v2/ spelling is tried first, got %q", paths[0])
			}
		})
	}
}

// includeDeleted defaults to TRUE on /v2/entities/list, the opposite of nearly
// everywhere else, so it is always sent explicitly.
func TestEntitiesListSendsIncludeDeletedExplicitly(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		incl bool
		want string
	}{
		{name: "excluded", incl: false, want: "false"},
		{name: "included", incl: true, want: "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotQ url.Values
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotQ = r.URL.Query()
				_, _ = w.Write([]byte(`[{"id":"e-1","name":"Наличные"}]`))
			})
			got, err := s.ListReferenceEntities(context.Background(), []string{"PaymentType", "Account"}, tt.incl, -1)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("decoded %d, want 1", len(got))
			}
			if gotQ.Get("includeDeleted") != tt.want {
				t.Errorf("includeDeleted = %q, want %q sent explicitly — it defaults to TRUE here", gotQ.Get("includeDeleted"), tt.want)
			}
			if len(gotQ["rootType"]) != 2 {
				t.Errorf("rootType must repeat per value, got %v", gotQ["rootType"])
			}
		})
	}
}

func TestReferenceSearchesHitTheirPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		call     func(*Service) error
		wantPath string
		wantQ    string
	}{
		{
			name: "departments", wantPath: EndpointDepartmentsSearch, wantQ: "code",
			call: func(s *Service) error { _, err := s.SearchDepartments(context.Background(), "D.*"); return err },
		},
		{
			name: "stores", wantPath: EndpointStoresSearch, wantQ: "code",
			call: func(s *Service) error { _, err := s.SearchStores(context.Background(), "S.*"); return err },
		},
		{
			name: "groups", wantPath: EndpointGroupsSearch, wantQ: "name",
			call: func(s *Service) error { _, err := s.SearchGroups(context.Background(), "Bar", ""); return err },
		},
		{
			name: "terminals", wantPath: EndpointTerminalsSearch, wantQ: "name",
			call: func(s *Service) error {
				_, err := s.SearchTerminals(context.Background(), map[string]string{"name": "POS"})
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			var gotQ url.Values
			s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotQ = r.URL.Path, r.URL.Query()
				_, _ = w.Write([]byte(`<corporateItemDtoes><corporateItemDto><id>d1</id></corporateItemDto></corporateItemDtoes>`))
			})
			if err := tt.call(s); err != nil {
				t.Fatalf("search: %v", err)
			}
			if !strings.HasSuffix(gotPath, tt.wantPath) {
				t.Errorf("path = %q, want suffix %q", gotPath, tt.wantPath)
			}
			if gotQ.Get(tt.wantQ) == "" {
				t.Errorf("%s must be sent, got %v", tt.wantQ, gotQ)
			}
		})
	}
}

// Replication is Chain-only and errors outright on an RMS, so ServerType is the
// capability probe callers are meant to gate on.
func TestReplicationIsGatedByServerType(t *testing.T) {
	t.Parallel()
	var gotPath string
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`[{"departmentId":"d-1","status":"OK"}]`))
	})
	if _, err := s.ReplicationStatuses(context.Background()); err != nil {
		t.Fatalf("statuses: %v", err)
	}
	if !strings.HasSuffix(gotPath, EndpointReplicationStatuses) {
		t.Errorf("path = %q", gotPath)
	}
	if _, err := s.ReplicationStatus(context.Background(), " "); err == nil {
		t.Error("a blank department id must be refused before the request")
	}
}

func TestEntityIDsPutsTheTypeInThePath(t *testing.T) {
	t.Parallel()
	var gotPath string
	var gotQ url.Values
	s, _ := newFake(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQ = r.URL.Path, r.URL.Query()
		_, _ = w.Write([]byte(`["id-1","id-2"]`))
	})
	got, err := s.EntityIDs(context.Background(), "Account", false, -1)
	if err != nil {
		t.Fatalf("ids: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("decoded %v", got)
	}
	if !strings.HasSuffix(gotPath, "/api/v2/entities/Account/ids") {
		t.Errorf("the entity type is a path segment, got %q", gotPath)
	}
	if gotQ.Get("includeDeleted") != "false" {
		t.Errorf("includeDeleted defaults TRUE here and must be explicit, got %q", gotQ.Get("includeDeleted"))
	}
}
